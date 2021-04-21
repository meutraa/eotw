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
	"git.lost.host/meutraa/eott/internal/score"
	"git.lost.host/meutraa/eott/internal/theme"
	"github.com/eiannone/keyboard"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"golang.org/x/term"
)

func main() {
	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

func isRowInField(rc int, row int, hit bool) bool {
	return !hit && (row < rc && row > 0)
}

func getColumn(nKeys uint8, mc, index int) int {
	return mc - int(nKeys>>1)*int(*config.ColumnSpacing*2) + index*int(*config.ColumnSpacing*2)
}

func judge(d float64) (int, *game.Judgement) {
	for i, j := range config.Judgements {
		if d < j.Ms {
			return i, &j
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
	var scorer score.Scorer = &score.DefaultScorer{}

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

	charts, err := psr.Parse(chartFile)
	if nil != err {
		return err
	}

	err = scorer.Init()
	if nil != err {
		return err
	}
	defer func() {
		scorer.Deinit()
	}()

	// Difficulty selection
	for i, c := range charts {
		histories := scorer.Load(c)
		fmt.Printf(
			"%2v) %v-key    %3v %v\n\tNotes: %5v\n",
			i,
			c.Difficulty.NKeys,
			c.Difficulty.Msd,
			c.Difficulty.Name,
			len(c.Notes),
		)
		for i, history := range histories {
			sc := scorer.Score(c, &history)
			fmt.Printf("\t\t%v: %2.1fx  Misses: %4v   Total Error: %v\n",
				i,
				history.Rate,
				sc.MissCount,
				sc.TotalError.Milliseconds(),
			)
		}
		fmt.Printf("\n")
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

	sideCol := getColumn(chart.Difficulty.NKeys, mc, 0) - 36
	if sideCol < 2 {
		sideCol = 2
	}

	score := 0.0
	counts := make([]int, len(config.Judgements))
	sumOfDistance := 0.0
	mean := 0.0
	totalHits := 0.0
	stdev := 0.0
	inputs := []game.Input{}

	finished := false

	go func() {
		time.Sleep(*config.Delay + *config.Offset)
		speaker.Play(streamer)
	}()

	r.RenderLoop(*config.Delay, func(now, deadline time.Time, duration time.Duration) bool {
		if scorer.Distance(*config.Rate, chart.Notes[len(chart.Notes)-1], duration.Milliseconds()) < 0 {
			finished = true
			return false
		}

		// get the key inputs that occured so far
		for i := 0; i < len(keyChannel); i++ {
			key := <-keyChannel
			if key.Key == keyboard.KeyEsc {
				return false
			}

			// Which index was hit
			noteCol := config.KeyColumn(key.Rune, chart.Difficulty.NKeys)
			if -1 == noteCol {
				continue // This is not a valid input
			}
			input := game.Input{Index: noteCol, HitTime: duration}
			inputs = append(inputs, input)

			scorer.ApplyInputToChart(chart, &input, *config.Rate, func(note *game.Note, distance, absDistance float64) {
				score += absDistance
				totalHits += 1
				sumOfDistance += distance
				// because distance is < missDistance, this should never be nil
				index, _ := judge(distance)
				counts[index]++
				if totalHits > 1 {
					stdev = 0.0
					mean = sumOfDistance / totalHits
					for _, n := range chart.Notes {
						if n.HitTime == 0 {
							continue
						}
						diff := scorer.Distance(*config.Rate, n, n.HitTime)
						xi := float64(diff) - mean
						xi2 := xi * xi
						stdev += xi2
					}
					stdev /= (totalHits - 1)
					stdev = math.Sqrt(stdev)
				}
			})

		}

		// Render the hit bar
		for i := 0; i < int(chart.Difficulty.NKeys); i++ {
			r.Fill(rc-int(*config.BarRow), getColumn(chart.Difficulty.NKeys, mc, i), th.RenderHitField(i))
		}

		// Render notes
		for _, note := range chart.Notes {
			// clear all existing renders
			col := getColumn(chart.Difficulty.NKeys, mc, note.Index)
			if isRowInField(rc, note.Row, false) {
				r.Fill(note.Row, col, " ")
			}

			// Calculate the new row based on time
			nr := rc - int(*config.BarRow)
			d := scorer.Distance(*config.Rate, note, duration.Milliseconds())
			distance := int(math.Round(float64(d) / config.ScrollSpeed))
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
		for i, judgement := range config.Judgements {
			r.Fill(18+i, sideCol, fmt.Sprintf("%v:  %6v", judgement.Name, counts[i]))
		}

		return true
	})

	if finished {
		scorer.Save(chart, &inputs, *config.Rate)
		log.Println("saved")
	}
	_ = <-keyChannel
	return nil
}
