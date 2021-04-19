package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
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
	speed  = 1000 / frameRate * 3
	barRow = 8 // from the bottom of the screen
)

var (
	keys       = [nKey]string{"_", "-", "m", "p"}
	judgements = []game.Judgement{
		{Ms: 5, Name: "      \033[1;31mE\033[38;5;208mx\033[1;33ma\033[1;32mc\033[38;5;153mt\033[0m"},
		{Ms: 10, Name: " \033[1;35mRidiculous\033[0m"},
		{Ms: 20, Name: "  \033[38;5;153mMarvelous\033[0m"},
		{Ms: 40, Name: "      \033[1;36mGreat\033[0m"},
		{Ms: 60, Name: "       \033[1;32mGood\033[0m"},
		{Ms: missDistance, Name: "       \033[1;31mOkay\033[0m"},
		{Ms: -1, Name: "       \033[1;31mMiss\033[0m"},
	}
)

func main() {
	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

func isRowInField(rc int, row int, hit bool) bool {
	return !hit && (row < rc && row > 0)
}

// Use the global offest to calculate the distance to a note
func getDistance(n *game.Note, duration time.Duration) int64 {
	return int64(math.Round(float64(n.Ms)/(*config.Rate))) - duration.Milliseconds() + *config.Offset
}

func judge(d float64) (int, *game.Judgement) {
	for i := 0; i < len(judgements)-1; i++ {
		judgement := judgements[i]
		if d < judgement.Ms {
			return i, &judgement
		}
	}
	// This should never happen, since a check for d < missDistance is made
	return -1, nil
}

func run() error {
	// Ensure our Default implementations are used as interfaces
	var r render.Renderer = &render.DefaultRenderer{}
	var th theme.Theme = &theme.DefaultTheme{}
	var psr parser.Parser = &parser.DefaultParser{}

	columns, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if nil != err {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}
	rc, cc := rows, columns

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

	if err := filepath.Walk(*config.Directory, func(p string, info os.FileInfo, err error) error {
		switch path.Ext(info.Name()) {
		case ".mp3":
			mp3File = p
		case ".ogg":
			ogg = p
		case ".sm":
			chartFile = p
		}
		return nil
	}); nil != err {
		return fmt.Errorf("unable to walk song directory: %w", err)
	}

	if (mp3File == "" && ogg == "") || chartFile == "" {
		return errors.New("unable to find .sm and .mp3/.ogg file in given directory")
	}

	mc := cc >> 1
	cen := rc >> 1
	cis := &([nKey]int{
		mc - *config.ColumnSpacing*3,
		mc - *config.ColumnSpacing,
		mc + *config.ColumnSpacing,
		mc + *config.ColumnSpacing*3,
	})
	sideCol := cis[0] - 36
	if sideCol < 2 {
		sideCol = 2
	}

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

	speaker.Init(beep.SampleRate(math.Round(float64(format.SampleRate)*(*config.Rate))), format.SampleRate.N(time.Second/60))

	// Clear the screen and hide the cursor
	r.Init()
	defer func() {
		// Restore the terminal state
		r.Deinit()
	}()

	go func() {
		time.Sleep(time.Duration(*config.Delay) * time.Millisecond)
		speaker.Play(streamer)
	}()

	score := 0.0
	counts := make([]int, len(judgements))
	sumOfDistance := 0.0
	mean := 0.0
	totalHits := 0.0
	stdev := 0.0

	r.RenderLoop(time.Duration(*config.Delay)*time.Millisecond, func(now, deadline time.Time, duration time.Duration) bool {
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
				if (note.HitTime != 0) ||
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
				index, _ := judge(distance)
				counts[index]++
				closestNote.HitTime = int64(math.Round(float64(duration.Milliseconds()) * *config.Rate))
				if totalHits > 1 {
					stdev = 0.0
					mean = sumOfDistance / totalHits
					for _, n := range chart.Notes {
						if n.HitTime == 0 {
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
			r.Fill(rc-barRow, cis[i], th.RenderHitField(i))
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
			distance := int(math.Round(float64(getDistance(note, duration))) / float64(speed))
			note.Row = nr - distance

			// Is this row within the playing field?
			if !note.Miss && note.Row > rc && note.HitTime == 0 && !note.IsMine {
				counts[len(counts)-1] += 1
				note.Miss = true
				r.AddDecoration(col-1, cen-1, "\033[1;31m╭", 240)
				r.AddDecoration(col+1, cen-1, "\033[1;31m╮", 240)
				r.AddDecoration(col-1, cen, "\033[1;31m╰", 240)
				r.AddDecoration(col+1, cen, "\033[1;31m╯", 240)
			} else if isRowInField(rc, note.Row, note.HitTime != 0) {
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
			r.Fill(18+i, sideCol, fmt.Sprintf("%v:  %6v", judgement.Name, counts[i]))
		}

		return true
	})
	_ = <-keyChannel
	return nil
}
