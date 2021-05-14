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
	"os/signal"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strconv"
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

	var totalFrameCounter, totalRenderDuration uint64 = 0, 0

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, os.Interrupt)

	// Get the dimensions of the terminal
	_columns, _rows, err := term.GetSize(int(os.Stdout.Fd()))
	if nil != err {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}
	totalRows, columnCount := uint16(_rows), uint16(_columns)
	middleColumn, middleRow := columnCount>>1, totalRows>>1
	hitRow := totalRows - *config.BarOffsetFromBottom

	// Start reading keyboard input
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

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

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
	sideColData := 14 + sideCol

	distanceError, sumOfDistance := time.Millisecond, time.Millisecond
	counts := make([]int, len(config.Judgements))
	var mean, stdev float64 = 0.0, 0.0
	var totalHits uint64 = 0
	inputs := []game.Input{}

	finished := false

	go func() {
		time.Sleep(*config.Delay + *config.Offset)
		speaker.Play(streamer)
		song := buffer.Streamer(0, buffer.Len())
		speaker.Play(song)
		for {
			if song.Position() == song.Len() {
				quitChannel <- os.Interrupt
			}
			time.Sleep(time.Second)
		}
	}()

	// Render the hit bar
	for i := uint8(0); i < chart.Difficulty.NKeys; i++ {
		r.Fill(totalRows-*config.BarOffsetFromBottom, getColumn(chart.Difficulty.NKeys, middleColumn, i), th.RenderHitField(i))
	}

	// Render the static stat ui
	r.Fill(3, sideCol, "     Render:  ")
	r.Fill(10, sideCol, "   Error dt:  ")
	r.Fill(11, sideCol, "      Stdev:  ")
	r.Fill(12, sideCol, "       Mean:  ")
	r.Fill(13, sideCol, fmt.Sprintf("      Total:  %6v", chart.NoteCount))
	r.Fill(14, sideCol, fmt.Sprintf("      Mines:  %6v", chart.MineCount))
	for i, judgement := range config.Judgements {
		r.Fill(uint16(18+i), sideCol, judgement.Name+":  ")
	}

	// I do not care about all the above, it is instantish
	if *config.CpuProfile != "" {
		f, err := os.Create(*config.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	r.RenderLoop(*config.Delay,
		func(startTime time.Time, duration time.Duration) bool {
			// This fails if the last note is not meant to be hit
			if len(quitChannel) != 0 {
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
				r.AddDecoration(col, totalRows-*config.BarOffsetFromBottom, "*", 120)

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

				// Update stats
				r.Fill(10, sideColData, fmt.Sprintf("%6.0f ms", float64(distanceError)/float64(time.Millisecond)))
				r.Fill(11, sideColData, fmt.Sprintf("%6.2f ms", stdev/float64(time.Millisecond)))
				r.Fill(12, sideColData, fmt.Sprintf("%6.2f ms", mean/float64(time.Millisecond)))
				r.Fill(uint16(18+idx), sideColData, fmt.Sprintf("%6v", counts[idx]))
			}

			// Adjust the active note range
			// The first time this is called, the active slice is empty
			// and start, end = 0, 0
			active, start, end := chart.Active()
			startOffset := 0
			endOffset := 0

			// Render notes
			for _, note := range active {
				col := getColumn(chart.Difficulty.NKeys, middleColumn, note.Index)

				// rowCount = 60 = bottom of screen
				// BarRow = 8 = 8 from bottom of screen
				// nr = hitRow

				// Calculate the new row based on time

				// This is the main use of the Distance function
				d := scorer.Distance(*config.Rate, note.Time, duration)

				rowOffsetFromHitRow := int64(float64(d) * config.NsToRow)

				// Check if this note will be rendered
				if rowOffsetFromHitRow > int64(hitRow) {
					// This is too far in the future and off the top of the screen
					log.Fatalln("Active note should not be active: top")
				} else if rowOffsetFromHitRow < -int64(*config.BarOffsetFromBottom) {
					// This is scrolled past the bottom of the screen

					// Check to see if the note was missed
					if note.HitTime == 0 && !note.IsMine {
						eidx := len(counts) - 1
						counts[eidx] += 1
						r.Fill(uint16(18+eidx), sideColData, fmt.Sprintf("%6v", counts[eidx]))
						r.AddDecoration(col-1, middleRow-1, "\033[1;31m╭", 240)
						r.AddDecoration(col+1, middleRow-1, "\033[1;31m╮", 240)
						r.AddDecoration(col-1, middleRow, "\033[1;31m╰", 240)
						r.AddDecoration(col+1, middleRow, "\033[1;31m╯", 240)
					}

					// Mark the active window to slide forward 1
					startOffset++
					// Mark the render loop to clear this note
					r.Fill(note.Row, col, " ")
					// TODO: probably do not need this anymore
					note.Row = math.MaxUint16
				} else {
					// This is still an active note
					renderRow := hitRow - uint16(rowOffsetFromHitRow)

					// Only if this has changed position do we clear and render anew
					if note.Row != renderRow && note.HitTime == 0 {
						// TODO: there might be an optimization here
						r.Fill(note.Row, col, " ")
						note.Row = renderRow
						if note.IsMine {
							r.Fill(note.Row, col, th.RenderMine(col, note.Denom))
						} else {
							r.Fill(note.Row, col, th.RenderNote(col, note.Denom))
						}
					}
				}

			}

			// At the end of this render loop I want to see which notes will require rendering next frame and slide the window
			for _, note := range chart.Notes[end:] {
				d := scorer.Distance(*config.Rate, note.Time, duration)
				rowOffsetFromHitRow := int64(float64(d) * config.NsToRow)

				// Check if this note will be rendered
				if rowOffsetFromHitRow < int64(hitRow) {
					endOffset++
				} else {
					break
				}
			}

			// Update the sliding window
			chart.SetActive(start+startOffset, end+endOffset)

			return true
		},
		func(renderDuration time.Duration) {
			totalRenderDuration += uint64(renderDuration)
			totalFrameCounter++
			if totalFrameCounter%uint64(*config.DebugUpdateRate) != 0 {
				return
			}

			// Print debugging stats
			active, start, end := chart.Active()
			r.Fill(2, sideCol, fmt.Sprintf("     Window:  %v - %v (%v)", start, end, len(active)))
			r.Fill(3, sideColData, strconv.FormatUint(totalRenderDuration/totalFrameCounter, 10)+" ")
		},
	)

	if finished && len(quitChannel) == 0 {
		scorer.Save(chart, &inputs, *config.Rate)
		log.Println("saved")
		_, _ = <-in, <-in
	}
	return nil
}
