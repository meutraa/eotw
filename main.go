package main

// #include <linux/input-event-codes.h>
// #include <linux/input.h>
import "C"

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
	"git.lost.host/meutraa/eott/internal/game"
	"git.lost.host/meutraa/eott/internal/input"
	"git.lost.host/meutraa/eott/internal/parser"
	"git.lost.host/meutraa/eott/internal/render"
	"git.lost.host/meutraa/eott/internal/score"
	"git.lost.host/meutraa/eott/internal/theme"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"golang.org/x/term"

	_ "net/http/pprof"
)

func main() {
	config.Init()
	if *config.CpuProfile != "" {
		f, err := os.Create(*config.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

func getColumn(nKeys uint8, mc uint16, index uint8) uint16 {
	// 4 => 2
	mid := nKeys >> 1
	if index < mid {
		return mc - *config.ColumnSpacing*uint16(nKeys-mid-index)
	} else {
		return mc + *config.ColumnSpacing*uint16(index-mid)
	}
}

func judge(d time.Duration) (int, *game.Judgement) {
	for i, j := range config.Judgements {
		if d < j.Time {
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

	_columns, _rows, err := term.GetSize(int(os.Stdout.Fd()))
	if nil != err {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}
	rowCount, columnCount := uint16(_rows), uint16(_columns)
	middleColumn, middleRow := columnCount>>1, rowCount>>1

	in := make(chan *input.Event, 128)
	input.ReadInput(*config.Input, in)

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
			fmt.Printf("\t\t%v: %v%%  Misses: %4v   Total Error: %v\n",
				i,
				history.Rate,
				sc.MissCount,
				sc.TotalError.Milliseconds(),
			)
		}
		fmt.Printf("\n")
	}
	// TODO key := <-in
	/*index, err := strconv.ParseInt(string(key.Rune), 10, 64)
	if nil != err || index > int64(len(charts)-1) {
		return err
	}*/
	index := 0

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

	speaker.Init(beep.SampleRate(math.Round(0.01*float64(format.SampleRate)*float64(*config.Rate))), format.SampleRate.N(time.Second/60))

	// Clear the screen and hide the cursor
	r.Init()
	defer func() {
		// Restore the terminal state
		r.Deinit()
	}()

	sideCol := getColumn(chart.Difficulty.NKeys, middleColumn, 0) - 36
	if sideCol < 2 {
		sideCol = 2
	}

	distanceError, sumOfDistance := time.Millisecond, time.Millisecond
	counts := make([]int, len(config.Judgements))
	var mean, stdev float64 = 0.0, 0.0
	var totalHits uint64 = 0
	inputs := []game.Input{}

	finished := false

	go func() {
		time.Sleep(*config.Delay + *config.Offset)
		speaker.Play(streamer)
	}()

	// Render the hit bar
	for i := uint8(0); i < chart.Difficulty.NKeys; i++ {
		r.Fill(rowCount-*config.BarRow, getColumn(chart.Difficulty.NKeys, middleColumn, i), th.RenderHitField(i))
	}

	lastNote := chart.Notes[len(chart.Notes)-1]
	r.RenderLoop(*config.Delay, func(startTime time.Time, duration time.Duration) bool {
		// This fails if the last note is not meant to be hit
		if lastNote.Miss || lastNote.HitTime != 0 {
			finished = true
			return false
		}

		// get the key inputs that occured so far
		for i := 0; i < len(in); i++ {
			ev := <-in
			if ev.Code == C.KEY_ESC {
				return false
			}

			// Which index was hit
			if ev.Released {
				continue
			}
			index, err := config.KeyColumn(ev.Code, chart.Difficulty.NKeys)
			if nil != err {
				continue
			}
			input := game.Input{
				Index:   index,
				HitTime: time.Unix(0, ev.Time.Nano()).Sub(startTime),
			}

			inputs = append(inputs, input)
			// Get the column to render the hit splash at
			col := getColumn(chart.Difficulty.NKeys, middleColumn, input.Index)
			r.AddDecoration(col, rowCount-*config.BarRow, "*", 120)

			note, distance, abs := scorer.ApplyInputToChart(chart, &input, *config.Rate)
			if note == nil {
				continue
			}

			distanceError += abs
			totalHits += 1
			sumOfDistance += distance
			// because distance is < missDistance, this should never be nil
			idx, judgement := judge(abs)
			r.AddDecoration(middleColumn, middleRow, judgement.Name, 120)

			counts[idx]++
			if totalHits > 1 {
				stdev = 0.0
				mean = float64(sumOfDistance) / float64(totalHits)
				for _, n := range chart.Notes {
					if n.HitTime == 0 {
						continue
					}
					diff := scorer.Distance(*config.Rate, n.Time, n.HitTime)
					xi := float64(diff) - mean
					xi2 := xi * xi
					stdev += xi2
				}
				stdev /= float64(totalHits - 1)
				stdev = math.Sqrt(stdev)
			}
		}

		// Render notes
		for _, note := range chart.Notes {
			// clear all existing renders
			col := getColumn(chart.Difficulty.NKeys, middleColumn, note.Index)
			if note.Visible() {
				r.Fill(note.Row, col, " ")
			}

			// Calculate the new row based on time
			nr := rowCount - *config.BarRow
			// This is the main use of the Distance function
			d := scorer.Distance(*config.Rate, note.Time, duration)
			row := int64(float64(d) * config.NsToRow)

			// Check if this note will be rendered
			if row > int64(rowCount-*config.BarRow) || row < -int64(*config.BarRow) {
				note.Row = math.MaxUint16
			} else {
				note.Row = nr - uint16(row)
			}

			// Is this row within the playing field?
			if !note.Miss && note.HitTime == 0 && !note.IsMine && d < -config.Judgements[len(config.Judgements)-2].Time {
				counts[len(counts)-1] += 1
				note.Miss = true
				r.AddDecoration(col-1, middleRow-1, "\033[1;31m╭", 240)
				r.AddDecoration(col+1, middleRow-1, "\033[1;31m╮", 240)
				r.AddDecoration(col-1, middleRow, "\033[1;31m╰", 240)
				r.AddDecoration(col+1, middleRow, "\033[1;31m╯", 240)
			} else if note.Visible() && note.HitTime == 0 {
				if note.IsMine {
					r.Fill(note.Row, col, th.RenderMine(col, note.Denom))
				} else {
					r.Fill(note.Row, col, th.RenderNote(col, note.Denom))
				}
			}
		}

		r.Fill(10, sideCol, fmt.Sprintf("   Error dt:  %6v", distanceError))
		r.Fill(11, sideCol, fmt.Sprintf("      Stdev:  %6.2f", stdev))
		r.Fill(12, sideCol, fmt.Sprintf("       Mean:  %6.2f", mean))
		r.Fill(13, sideCol, fmt.Sprintf("      Total:  %6v", chart.NoteCount))
		r.Fill(14, sideCol, fmt.Sprintf("      Mines:  %6v", chart.MineCount))
		for i, judgement := range config.Judgements {
			r.Fill(uint16(18+i), sideCol, fmt.Sprintf("%v:  %6v", judgement.Name, counts[i]))
		}

		return true
	})

	if finished {
		scorer.Save(chart, &inputs, *config.Rate)
		log.Println("saved")
	}
	log.Println("Called distance:", score.DistanceCount)
	_, _ = <-in, <-in
	return nil
}
