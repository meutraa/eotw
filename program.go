package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"time"

	"git.lost.host/meutraa/eotw/internal/config"
	"git.lost.host/meutraa/eotw/internal/game"
	"git.lost.host/meutraa/eotw/internal/parser"
	"git.lost.host/meutraa/eotw/internal/score"
	"git.lost.host/meutraa/eotw/internal/theme"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Position struct {
	X, Y int32
}

type Decoration struct {
	frames          int
	key             int32
	note            *game.Note
	startedCounting bool
	startCounting   func(note *game.Note, key int32) bool
	render          func(remaining int)
}

type Program struct {
	Parser *parser.DefaultParser
	Scorer *score.DefaultScorer
	Theme  *theme.DefaultTheme

	startTime time.Time

	frameCounter  uint64
	width, height int32
	middle        Position
	hitRow        int32

	decorations []*Decoration

	audioFile, chartFile string
	music                *rl.Music
	musicLength          float32 // In seconds

	charts []*game.Chart
	chart  game.Chart

	sideCol int32

	// Stats for current chart
	distanceError, sumOfDistance time.Duration
	counts                       []int
	mean, stdev                  float64
	totalHits                    uint64
	inputs                       []game.Input
}

func (p *Program) Resize() {
	log.Println(rl.GetScreenHeight(), rl.GetScreenWidth())
	p.width = int32(rl.GetScreenWidth())
	p.height = int32(rl.GetScreenHeight())
	p.middle = Position{X: p.width / 2, Y: p.height / 2}
	p.hitRow = p.height - *config.BarOffsetFromBottom

	p.sideCol = getColumn(p.chart.Difficulty.NKeys, p.middle.X, 0) - 360
	if p.sideCol < 20 {
		p.sideCol = 20
	}
}

func (g *Program) Init() error {
	// Ensure our Default implementations are used as interfaces
	g.Parser = &parser.DefaultParser{}
	g.Scorer = &score.DefaultScorer{}
	g.Theme = &theme.DefaultTheme{}

	if err := filepath.Walk(*config.Directory, func(p string, info os.FileInfo, err error) error {
		switch path.Ext(info.Name()) {
		case ".ogg", ".mp3", ".xm", ".mod", ".wav":
			g.audioFile = p
		case ".sm":
			g.chartFile = p
		}
		return nil
	}); nil != err {
		return fmt.Errorf("unable to walk song directory: %w", err)
	}

	if (g.audioFile == "") || g.chartFile == "" {
		return errors.New("unable to find .sm and .mp3/.ogg file in given directory")
	}

	var err error
	g.charts, err = g.Parser.Parse(g.chartFile)
	if nil != err {
		return err
	}

	err = g.Scorer.Init()
	if nil != err {
		return err
	}
	defer func() {
		g.Scorer.Deinit()
	}()

	g.chart = *g.charts[0]
	g.counts = make([]int, len(config.Judgements))
	g.inputs = []game.Input{}

	g.Resize()

	return nil
}

func (p *Program) Update(duration time.Duration) {
	// get the key inputs that occured so far
	for key := rl.GetKeyPressed(); key != 0; key = rl.GetKeyPressed() {
		index, err := config.KeyColumn(key, p.chart.Difficulty.NKeys)
		if nil != err {
			log.Println("not a column index pressed")
			continue
		}
		input := game.Input{Index: index, HitTime: duration}

		p.inputs = append(p.inputs, input)

		// Get the column to render the hit splash at
		col := getColumn(p.chart.Difficulty.NKeys, p.middle.X, input.Index)

		note, distance, abs := p.Scorer.ApplyInputToChart(&p.chart, &input, *config.Rate)
		if note == nil {
			// If this is hitting nothing
			p.decorations = append(p.decorations, &Decoration{
				frames: 24,
				key:    key,
				note:   note,
				startCounting: func(note *game.Note, key int32) bool {
					return rl.IsKeyReleased(key)
				},
				render: func(remaining int) {
					g := rl.Gray
					g.A = uint8(float32(255) * (float32(remaining) / 24))
					rl.DrawCircle(col, p.hitRow, *config.NoteRadius-4, g)
				},
			})
			continue
		}

		p.distanceError += abs
		p.totalHits += 1
		p.sumOfDistance += distance
		// because distance is < missDistance, this should never be nil
		idx, judgement := judge(abs)

		p.decorations = append(p.decorations, &Decoration{
			frames: 24,
			key:    key,
			note:   note,
			startCounting: func(note *game.Note, key int32) bool {
				released := rl.IsKeyReleased(key)
				if released {
					note.ReleaseTime = duration
					return true
				}
				return false
			},
			render: func(remaining int) {
				g := judgement.Color
				gr := rl.Gray
				gr.A = uint8(float32(255) * (float32(remaining) / 24))
				rl.DrawCircle(col, p.hitRow, *config.NoteRadius, g)
				rl.DrawCircle(col, p.hitRow, *config.NoteRadius-2, rl.Black)
				rl.DrawCircle(col, p.hitRow, *config.NoteRadius-4, gr)
			},
		})

		os := int32(2*-distance.Milliseconds()) + p.middle.X
		p.decorations = append(p.decorations, &Decoration{
			frames: 120,
			render: func(remaining int) {
				g := judgement.Color
				g.A = uint8(float32(255) * (float32(remaining) / 120))
				rl.DrawRectangle(
					os-2,
					int32(float32(p.middle.Y)*1.2),
					4,
					20,
					g,
				)
			},
		})

		p.counts[idx]++
		if p.totalHits > 1 {
			p.stdev = 0.0
			p.mean = float64(p.sumOfDistance) / float64(p.totalHits)
			for _, n := range p.chart.Notes {
				if n.HitTime == 0 {
					continue
				}
				diff := p.Scorer.Distance(*config.Rate, n.Time, n.HitTime)
				xi := float64(diff) - p.mean
				xi2 := xi * xi
				p.stdev += xi2
			}
			p.stdev /= float64(p.totalHits - 1)
			p.stdev = math.Sqrt(p.stdev)
		}
	}
}

func (p *Program) Render(duration time.Duration) {
	p.frameCounter++

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	p.RenderBackgroundDecoration()
	p.RenderStatic()
	p.RenderGame(duration)

	rl.EndDrawing()
}

func (p *Program) RenderBackgroundDecoration() {
	// This might get big, but I think it is really fast
	for _, dec := range p.decorations {
		if dec.frames > 0 {
			if dec.startCounting == nil ||
				dec.startedCounting ||
				dec.startCounting(dec.note, dec.key) {
				dec.startedCounting = true
				dec.frames--
			}
			dec.render(dec.frames)
		}
	}
}

func pixelsFromHitbar(timeFromHitbar time.Duration) int64 {
	return int64(float64(timeFromHitbar) * config.PixelsPerNs)
}

func (p *Program) RenderGame(duration time.Duration) {
	// Adjust the active note range
	// The first time this is called, the active slice is empty
	// and start, end = 0, 0
	active, start, end := p.chart.Active()
	startOffset := 0
	endOffset := 0

	// Render notes
	for _, note := range active {
		col := getColumn(p.chart.Difficulty.NKeys, p.middle.X, note.Index)

		// This is the main use of the Distance function
		d := p.Scorer.Distance(*config.Rate, note.Time, duration)

		worst := config.Judgements[len(config.Judgements)-2]

		// Check if this note will be rendered
		if d < -worst.Time {
			// This is scrolled past the bottom of the screen
			// Check to see if the note was missed

			if note.HitTime == 0 && note.MissTime == 0 && !note.IsMine {
				eidx := len(p.counts) - 1
				note.MissTime = duration
				p.counts[eidx] += 1
				os := int32(2*-worst.Time.Milliseconds()) + p.middle.X
				p.decorations = append(p.decorations, &Decoration{
					frames: 120,
					render: func(remaining int) {
						g := config.Judgements[len(config.Judgements)-1].Color
						g.A = uint8(float32(255) * (float32(remaining) / 120))
						rl.DrawRectangle(
							os-3,
							int32(float32(p.middle.Y)*1.2)-5,
							6,
							30,
							g,
						)
					},
				})
			}

			// Mark the active window to slide forward 1
			// First check if we should keep rendering for the sake
			// of holds spanning the entire screen
			if active[0].TimeEnd != 0 {
				de := p.Scorer.Distance(*config.Rate, note.TimeEnd, duration)
				if de < -worst.Time {
					startOffset++
				} // else holding window because end of hold note still active
			} else {
				startOffset++
			}
		}

		if (note.HitTime == 0 && note.TimeEnd == 0) || (note.TimeEnd != 0) {
			// This is still an active, relevant note
			ps := pixelsFromHitbar(d)
			x, y := col, p.hitRow-int32(ps)

			if note.IsMine {
				rl.DrawCircleLines(x, y, *config.NoteRadius, rl.DarkGray)
			} else {
				r, g, b := p.Theme.GetNoteColor(note.Denom)
				color := rl.NewColor(r, g, b, 255)

				if note.TimeEnd != 0 {
					// This is a hold note
					de := p.Scorer.Distance(*config.Rate, note.TimeEnd, duration)
					pe := pixelsFromHitbar(de)
					ye := p.hitRow - int32(pe)
					if note.MissTime != 0 {
						// 250ms until gone
						timeSince := duration.Milliseconds() - note.MissTime.Milliseconds()
						alpha := timeSince * 3
						if alpha > 255 {
							color.A = 0
						} else {
							color.A = uint8(255 - alpha)
						}
					} else if note.HitTime != 0 {
						// fill from the hit time to end time
						dsh := p.Scorer.Distance(*config.Rate, note.HitTime, duration)
						psh := pixelsFromHitbar(dsh)

						_, judgement := judge(time.Duration(math.Abs(float64(note.Time - note.HitTime))))

						if note.ReleaseTime != 0 {
							deh := p.Scorer.Distance(*config.Rate, note.ReleaseTime, duration)
							peh := pixelsFromHitbar(deh)
							yeh := p.hitRow - int32(psh)

							rl.DrawRectangleRounded(
								rl.Rectangle{
									X:      float32(col) - *config.NoteRadius + 2,
									Y:      float32(yeh) - *config.NoteRadius + 2,
									Width:  *config.NoteRadius*2 - 4,
									Height: float32(peh-psh) - 4,
								},
								1, 1, judgement.Color,
							)
						} else {
							yaeh := p.hitRow
							if yaeh < ye {
								yaeh = ye
							}
							rl.DrawRectangleRounded(
								rl.Rectangle{
									X:      float32(col) - *config.NoteRadius + 2,
									Y:      float32(yaeh) - *config.NoteRadius + 2,
									Width:  *config.NoteRadius*2 - 4,
									Height: float32(-psh) - 4,
								},
								1, 1, judgement.Color,
							)
							// fill from hit time to current time
						}
					}
					rl.DrawRectangleRoundedLines(
						rl.Rectangle{
							X:      float32(col) - *config.NoteRadius,
							Y:      float32(ye) - *config.NoteRadius,
							Width:  *config.NoteRadius * 2,
							Height: float32(pe - ps),
						},
						1, 1, 2, color,
					)

				} else {
					rl.DrawCircle(x, y, *config.NoteRadius, color)
				}
			}
		}
	}

	// At the end of this render loop I want to see which notes will require rendering
	// next frame and slide the window
	for _, note := range p.chart.Notes[end:] {
		d := p.Scorer.Distance(*config.Rate, note.Time, duration)

		// Check if this note will be rendered
		if pixelsFromHitbar(d) < int64(p.hitRow) {
			endOffset++
		} else {
			break
		}
	}

	// Update the sliding window
	p.chart.SetActive(start+startOffset, end+endOffset)
}

func (p *Program) RenderStatic() {
	// Render the hit bar
	for i := uint8(0); i < p.chart.Difficulty.NKeys; i++ {
		g := rl.Gray
		g.A = 128
		rl.DrawCircleLines(
			getColumn(p.chart.Difficulty.NKeys, p.middle.X, i),
			p.hitRow,
			*config.NoteRadius+4,
			g,
		)
	}

	rl.DrawRectangle(0, 2, int32(float32(p.width)*(rl.GetMusicTimePlayed(*p.music)/p.musicLength)), 2, rl.White)

	// Render the static stat ui
	rl.DrawFPS(p.sideCol, 3*24)
	notes, start, end := p.chart.Active()
	rl.DrawText(fmt.Sprintf(" Active Window [%v - %v] (%v)", start, end, len(notes)), p.sideCol, 4*24, 20, rl.White)
	rl.DrawText(fmt.Sprintf("   Error dt: %6.0f ms", float64(p.distanceError)/float64(time.Millisecond)), p.sideCol, 10*24, 20, rl.White)
	rl.DrawText(fmt.Sprintf("      Stdev: %6.2f ms", p.stdev/float64(time.Millisecond)), p.sideCol, 11*24, 20, rl.White)
	rl.DrawText(fmt.Sprintf("       Mean: %6.2f ms", p.mean/float64(time.Millisecond)), p.sideCol, 12*24, 20, rl.White)
	rl.DrawText(fmt.Sprintf("      Notes:  %6v", p.chart.NoteCount), p.sideCol, 13*24, 20, rl.White)
	rl.DrawText(fmt.Sprintf("      Holds:  %6v", p.chart.HoldCount), p.sideCol, 14*24, 20, rl.White)
	rl.DrawText(fmt.Sprintf("      Mines:  %6v", p.chart.MineCount), p.sideCol, 15*24, 20, rl.White)
	sh := int32(float32(p.middle.Y) * 1.2)
	for i, j := range config.Judgements {
		if i < len(config.Judgements)-1 {
			os := int32(2*-j.Time.Milliseconds()) + p.middle.X
			osp := int32(2*+j.Time.Milliseconds()) + p.middle.X
			col := j.Color
			col.A = 128
			rl.DrawLine(os, sh+5, os, sh+10, col)
			rl.DrawLine(osp, sh+5, osp, sh+10, col)
		}
		rl.DrawText(fmt.Sprintf("%s: %6v", j.Name, p.counts[i]), p.sideCol, int32(24*(18+i)), 20, j.Color)
	}
}
