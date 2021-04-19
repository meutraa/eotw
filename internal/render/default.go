package render

import (
	"fmt"
	"image"
	"image/color"
	"time"
)

const framePeriod = 1 * time.Millisecond // game loop/render deadline

type DefaultRenderer struct {
	buffer      string
	decorations []*Decoration
}

// Put the terminal in alt mode and clear the screen
func (r *DefaultRenderer) Init() {
	fmt.Printf("\033[?1049h\033[?25l\033[H\033[J")
}

// Restore the state of the terminal
func (r *DefaultRenderer) Deinit() {
	fmt.Printf("\033[?1049l\033[?25h")
}

func (r *DefaultRenderer) AddDecoration(col, row int, content string, frames int) {
	r.decorations = append(r.decorations, &Decoration{
		Point: image.Point{
			X: col,
			Y: row,
		},
		Content: content,
		Frames:  frames,
	})
	r.Fill(row, col, content)
}

func (r *DefaultRenderer) tickDecorations() {
	nd := []*Decoration{}
	for _, d := range r.decorations {
		if d.Frames == 0 {
			r.Fill(d.Point.Y, d.Point.X, " ")
			continue
		}
		nd = append(nd, d)
		d.Frames--
	}
	r.decorations = nd
}

func (r *DefaultRenderer) RenderLoop(delay time.Duration, render func(now, deadline time.Time, duration time.Duration) bool) {
	cont := true
	startTime := time.Now().Add(delay)
	for cont {
		now := time.Now()
		duration := now.Sub(startTime)
		deadline := now.Add(framePeriod)

		cont = render(now, deadline, duration)

		r.tickDecorations()
		r.flush()

		remainingTime := deadline.Sub(time.Now())
		// renderTime := framePeriod.Nanoseconds() - remainingTime.Nanoseconds()
		time.Sleep(remainingTime)
	}
}

func (r *DefaultRenderer) Fill(row, column int, message string) {
	r.buffer += fmt.Sprintf("\033[%d;%dH%v", row, column, message)
}

func (r *DefaultRenderer) FillColor(row, column int, c color.RGBA, message string) {
	r.buffer += fmt.Sprintf("\033[%d;%dH\033[38;2;%v;%v;%vm%v\033[0m", row, column, c.R, c.G, c.B, message)
}

func (r *DefaultRenderer) flush() {
	fmt.Print(r.buffer)
	r.buffer = ""
}
