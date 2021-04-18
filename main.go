package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.lost.host/meutraa/eott/internal/game"
	"git.lost.host/meutraa/eott/internal/graphics"
	"git.lost.host/meutraa/eott/internal/parser"
	"git.lost.host/meutraa/eott/internal/render"
	"git.lost.host/meutraa/eott/internal/theme"
	"github.com/eiannone/keyboard"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"golang.org/x/term"
)

const (
	globalOffset = 0.00
	frameRate    = 240
	missDistance = barRow * speed // I count off the screen as missed
	startDelay   = 3000 * time.Millisecond
	flashLength  = 60 // error flash length ms
	nKey         = 4
	// If the note is 300ms away, base distance is 300 rows, this divides that
	speed  = 1000 / frameRate * 3
	barRow = 8 // from the bottom of the screen
)

var (
	keys       = [...]string{"_", "-", "m", "p"}
	judgements = []Judgement{
		{0, 5, "      \033[1;31mE\033[38;5;208mx\033[1;33ma\033[1;32mc\033[38;5;153mt\033[0m"},
		{1, 10, " \033[1;35mRidiculous\033[0m"},
		{2, 20, "  \033[38;5;153mMarvelous\033[0m"},
		{3, 40, "      \033[1;36mGreat\033[0m"},
		{4, 60, "       \033[1;32mOkay\033[0m"},
		{5, missDistance, "       \033[1;31mMiss\033[0m"},
	}
)

type Judgement struct {
	index int
	ms    float64
	name  string
}

var song = flag.String("f", "", "path to dir containing mp3 & sm file")

func main() {
	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

func isRowInField(rc int64, row int64, hit bool) bool {
	return !hit && (row < rc && row > 0)
}

func step(r render.Renderer, th theme.Theme, note *game.Note, now time.Time, currentDuration time.Duration, cis *[4]int64, rc int64) (miss int) {
	// clear all existing renders
	col := cis[note.Index]
	if isRowInField(rc, note.Row, false) {
		r.Fill(note.Row, col, " ")
	}
	if note.MissFlashRemaining > 1 {
		note.MissFlashRemaining--
	} else if note.MissFlashRemaining == 1 {
		note.MissFlashRemaining--
		// clear flash
		for _, note := range *note.MissFlash {
			r.Fill(note.Row, note.Column, " ")
		}
	} // else 0 and does not exist anymore

	// Calculate the new row based on time
	nr := (rc - barRow)
	distance := int64(math.Round(float64((note.Ms - currentDuration.Milliseconds())) / float64(speed)))
	note.Row = nr - distance

	// Is this row within the playing field?
	if !note.Miss && note.Row > rc && !note.Hit && !note.IsMine {
		miss = 1
		note.Miss = true
		// flash the miss
		note.MissFlashRemaining = flashLength
		cen := rc >> 1
		note.MissFlash = &map[string]graphics.Point{
			"╭": {col - 1, cen - 1},
			"╮": {col + 1, cen - 1},
			"╰": {col - 1, cen},
			"╯": {col + 1, cen},
		}
		for c, p := range *note.MissFlash {
			r.Fill(p.Row, p.Column, "\033[1;31m"+c)
		}
	} else if isRowInField(rc, note.Row, note.Hit) {
		if note.IsMine {
			r.Fill(note.Row, col, th.RenderMine(note.Index, note.Denom))
		} else {
			r.Fill(note.Row, col, th.RenderNote(note.Index, note.Denom))
		}
	}
	return
}

// Returns an index of the judgement array
func judge(d float64) Judgement {
	for _, judgement := range judgements {
		if d < judgement.ms {
			return judgement
		}
	}
	return judgements[len(judgements)-1]
}

func run() error {
	// Ensure our Default implementations are used as interfaces
	var r render.Renderer = &render.DefaultRenderer{}
	var th theme.Theme = &theme.DefaultTheme{}
	var psr parser.Parser = &parser.DefaultParser{}

	flag.Parse()

	cc, rc, err := term.GetSize(int(os.Stdout.Fd()))
	if nil != err {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}

	keyChannel, err := keyboard.GetKeys(128)
	if nil != err {
		return fmt.Errorf("unable to open keyboard: %w", err)
	}
	defer func() {
		if err := keyboard.Close(); nil != err {
			log.Println("unable to close keyboard %w", err)
		}
	}()

	var mp3File, ogg, chartFile string

	if err := filepath.Walk(*song, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".mp3") {
			mp3File = path
		} else if strings.HasSuffix(info.Name(), ".ogg") {
			ogg = path
		} else if strings.HasSuffix(info.Name(), ".sm") {
			chartFile = path
		}
		return nil
	}); nil != err {
		return fmt.Errorf("unable to walk song directory: %w", err)
	}

	if (mp3File == "" && ogg == "") || chartFile == "" {
		return errors.New("unable to find .sm and .mp3 file in given directory")
	}

	mc := int64(cc) >> 1
	space := int64(6)
	cis := &([4]int64{mc - space*3, mc - space, mc + space, mc + space*3})
	sideCol := mc - space*9

	charts, err := psr.Parse(chartFile)
	if nil != err {
		return err
	}

	// Difficulty selection
	for i, c := range charts {
		fmt.Printf("%2v) %3v  %5v  %v\n", i, c.Difficulty.Msd, len(c.Notes), c.Difficulty.Name)
	}
	key := <-keyChannel
	index, err := strconv.ParseInt(string(key.Rune), 10, 64)
	if nil != err || index > int64(len(charts)-1) {
		return err
	}

	chart := charts[index]

	audioFile := mp3File
	if ogg != "" {
		audioFile = ogg
	}
	log.Printf("Opening %v (%v)\n", audioFile, chartFile)
	f, err := os.Open(audioFile)
	if err != nil {
		return err
	}
	var streamer beep.StreamSeekCloser
	var format beep.Format
	if ogg != "" {
		streamer, format, err = vorbis.Decode(f)
	} else {
		streamer, format, err = mp3.Decode(f)
	}
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/60))

	// Clear the screen and hide the cursor
	r.Init()
	defer func() {
		// Restore the terminal state
		r.Deinit()
	}()

	go func() {
		time.Sleep(startDelay)
		speaker.Play(streamer)
	}()

	score := 0.0
	counts := make([]int, len(judgements))

	// stdev
	sumOfDistance := 0.0
	mean := 0.0
	totalHits := 0.0
	stdev := 0.0

	r.RenderLoop(startDelay, func(now, deadline time.Time, duration time.Duration) bool {
		if duration.Milliseconds()-5000 > chart.Notes[len(chart.Notes)-1].Ms {
			return false
		}

		// get the key inputs that occured so far
		for i := 0; i < len(keyChannel); i++ {
			key := <-keyChannel
			if key.Key == keyboard.KeyEsc {
				return false
			}
			var closestNote *game.Note
			distance := 10000000.0
			dirDistance := 1000000.0
			for _, note := range chart.Notes {
				if (note.Hit) ||
					(note.IsMine) ||
					(note.Index != 0 && string(key.Rune) == keys[0]) ||
					(note.Index != 1 && string(key.Rune) == keys[1]) ||
					(note.Index != 2 && string(key.Rune) == keys[2]) ||
					(note.Index != 3 && string(key.Rune) == keys[3]) {
					continue
				}
				dd := float64(note.Ms - duration.Milliseconds())
				d := math.Abs(dd)
				// log.Printf("hiiiiiit %v %v = %v", note.ms, currentDuration.Milliseconds(), dd)
				if d < distance {
					dirDistance = dd
					distance = d
					closestNote = note
				} else if nil != closestNote {
					// already found the closest, and this d is > md
					break
				}
			}
			if nil != closestNote && distance < missDistance {
				score += distance
				totalHits += 1
				sumOfDistance += dirDistance
				judgement := judge(distance)
				counts[judgement.index]++
				closestNote.HitTime = duration.Milliseconds()
				closestNote.Hit = true
				if totalHits > 2 {
					stdev = 0.0
					mean = sumOfDistance / totalHits
					for _, n := range chart.Notes {
						if !n.Hit {
							continue
						}
						diff := float64(n.Ms - n.HitTime)
						xi := diff - mean
						xi2 := xi * xi
						stdev += xi2
					}
					stdev /= (totalHits - 1)
					stdev = math.Sqrt(stdev)
				}
			}
		}

		// Render the hit bar
		for i := 0; i < nKey; i++ {
			r.Fill(int64(rc)-barRow, cis[i], th.RenderHitField(i))
		}

		// Render notes
		for _, note := range chart.Notes {
			miss := step(r, th, note, now, duration, cis, int64(rc))
			counts[len(counts)-1] += miss
		}

		// remainingTime := deadline.Sub(time.Now())
		// renderTime := framePeriod.Nanoseconds() - remainingTime.Nanoseconds()

		// r.Fill(2, sideCol, fmt.Sprintf("Render Time:  %5.0f µs", float64(renderTime)/1000.0))
		// r.Fill(3, sideCol, fmt.Sprintf("  Idle Time:  %.1f%%", 100-100*float64(renderTime)/float64(framePeriod.Nanoseconds())))
		r.Fill(10, sideCol, fmt.Sprintf("   Error dt:  %6v", score))
		r.Fill(11, sideCol, fmt.Sprintf("      Stdev:  %6.2f", stdev))
		r.Fill(12, sideCol, fmt.Sprintf("       Mean:  %6.2f", mean))
		r.Fill(13, sideCol, fmt.Sprintf("      Total:  %6v", chart.NoteCount))
		r.Fill(14, sideCol, fmt.Sprintf("      Mines:  %6v", chart.MineCount))
		for i, judgement := range judgements {
			r.Fill(int64(18+i), sideCol, fmt.Sprintf("%v:  %6v", judgement.name, counts[i]))
		}
		r.Flush()

		return true
	})
	_ = <-keyChannel
	return nil
}
