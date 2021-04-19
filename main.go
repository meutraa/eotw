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
	frameRate    = 240
	missDistance = barRow * speed // I count off the screen as missed
	nKey         = 4
	// If the note is 300ms away, base distance is 300 rows, this divides that
	speed         = 1000 / frameRate * 3
	barRow        = 8 // from the bottom of the screen
	columnSpacing = 6 // columns between the note columns
)

var (
	song         = flag.String("f", "", "path to dir containing mp3 & sm file")
	rate         = flag.Float64("r", 1.0, "playback seed")
	globalOffset = flag.Int64("o", 0, "gloabl offset (milliseconds)")
	startDelay   = flag.Int64("d", 1500, "start delay (milliseconds)")

	keys       = [nKey]string{"_", "-", "m", "p"}
	judgements = []game.Judgement{
		{0, 5, "      \033[1;31mE\033[38;5;208mx\033[1;33ma\033[1;32mc\033[38;5;153mt\033[0m"},
		{1, 10, " \033[1;35mRidiculous\033[0m"},
		{2, 20, "  \033[38;5;153mMarvelous\033[0m"},
		{3, 40, "      \033[1;36mGreat\033[0m"},
		{4, 60, "       \033[1;32mGood\033[0m"},
		{5, missDistance, "       \033[1;31mOkay\033[0m"},
		{6, -1, "       \033[1;31mMiss\033[0m"},
		// > miss = not hit at all
	}
)

func main() {
	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

func isRowInField(rc int64, row int64, hit bool) bool {
	return !hit && (row < rc && row > 0)
}

// Use the global offest to calculate the distance to a note
func getDistance(n *game.Note, duration time.Duration) int64 {
	return int64(math.Round(float64(n.Ms)/(*rate))) - duration.Milliseconds() + *globalOffset
}

func judge(d float64) *game.Judgement {
	for i := 0; i < len(judgements)-1; i++ {
		judgement := judgements[i]
		if d < judgement.Ms {
			return &judgement
		}
	}
	// This should never happen, since a check for d < missDistance is made
	return nil
}

func run() error {
	// Ensure our Default implementations are used as interfaces
	var r render.Renderer = &render.DefaultRenderer{}
	var th theme.Theme = &theme.DefaultTheme{}
	var psr parser.Parser = &parser.DefaultParser{}

	flag.Parse()

	columns, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if nil != err {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}
	rc, cc := int64(rows), int64(columns)

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
	cen := rc >> 1
	cis := &([nKey]int64{
		mc - columnSpacing*3,
		mc - columnSpacing,
		mc + columnSpacing,
		mc + columnSpacing*3,
	})
	sideCol := mc - columnSpacing*9

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

	speaker.Init(beep.SampleRate(math.Round(float64(format.SampleRate)*(*rate))), format.SampleRate.N(time.Second/60))

	// Clear the screen and hide the cursor
	r.Init()
	defer func() {
		// Restore the terminal state
		r.Deinit()
	}()

	go func() {
		time.Sleep(time.Duration(*startDelay) * time.Millisecond)
		speaker.Play(streamer)
	}()

	score := 0.0
	counts := make([]int, len(judgements))

	// stdev
	sumOfDistance := 0.0
	mean := 0.0
	totalHits := 0.0
	stdev := 0.0

	/*{
		cd := 96 / (len(judgements) - 1)
		for i := 1; i < len(judgements)-1; i++ {
			j := judgements[i]
			bd := getDistance(&game.Note{Ms: int64(j.Ms)}, 0)
			bdd := int64(math.Round(float64(bd)) / float64(speed))
			c := 156 - cd*(i+1)
			cm1 := c - (cd / 2)
			cm2 := c - (cd/2)*2
			cm3 := c - (cd/2)*3
			cm4 := c - (cd/2)*4
			m := fmt.Sprintf("\033[38;2;%v;%v;%vm───", c, c, c)
			m1 := fmt.Sprintf("\033[38;2;%v;%v;%vm─", cm1, cm1, cm1)
			m2 := fmt.Sprintf("\033[38;2;%v;%v;%vm─", cm2, cm2, cm2)
			m3 := fmt.Sprintf("\033[38;2;%v;%v;%vm─", cm3, cm3, cm3)
			m4 := fmt.Sprintf("\033[38;2;%v;%v;%vm─", cm4, cm4, cm4)
			r.AddDecoration(2, rc-barRow-bdd, fmt.Sprintf("%v%v%v%v%v%v%v%v%v\033[0m", m4, m3, m2, m1, m, m1, m2, m3, m4), 24000)
			r.AddDecoration(2, rc-barRow+bdd, fmt.Sprintf("%v%v%v%v%v%v%v%v%v\033[0m", m4, m3, m2, m1, m, m1, m2, m3, m4), 24000)
		}
	}*/

	r.RenderLoop(time.Duration(*startDelay)*time.Millisecond, func(now, deadline time.Time, duration time.Duration) bool {
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
				dd := getDistance(note, duration)
				d := math.Abs(float64(dd))
				if d < distance {
					dirDistance = float64(dd)
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
				// because distance is < missDistance, this should never be nil
				judgement := judge(distance)
				counts[judgement.Index]++
				closestNote.HitTime = duration.Milliseconds()
				closestNote.Hit = true
				if totalHits > 1 {
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
			// clear all existing renders
			col := cis[note.Index]
			if isRowInField(rc, note.Row, false) {
				r.Fill(note.Row, col, " ")
			}

			// Calculate the new row based on time
			nr := (rc - barRow)
			distance := int64(math.Round(float64(getDistance(note, duration))) / float64(speed))
			note.Row = nr - distance

			// Is this row within the playing field?
			if !note.Miss && note.Row > rc && !note.Hit && !note.IsMine {
				counts[len(counts)-1] += 1
				note.Miss = true
				r.AddDecoration(col-1, cen-1, "\033[1;31m╭", 240)
				r.AddDecoration(col+1, cen-1, "\033[1;31m╮", 240)
				r.AddDecoration(col-1, cen, "\033[1;31m╰", 240)
				r.AddDecoration(col+1, cen, "\033[1;31m╯", 240)
			} else if isRowInField(rc, note.Row, note.Hit) {
				if note.IsMine {
					r.Fill(note.Row, col, th.RenderMine(note.Index, note.Denom))
				} else {
					r.Fill(note.Row, col, th.RenderNote(note.Index, note.Denom))
				}
			}
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
			r.Fill(int64(18+i), sideCol, fmt.Sprintf("%v:  %6v", judgement.Name, counts[i]))
		}

		return true
	})
	_ = <-keyChannel
	return nil
}
