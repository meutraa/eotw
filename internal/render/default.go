package render

import (
	"fmt"
	"image/color"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
)

type DefaultRenderer struct {
	buffer      string
	decorations []*decoration
}

type decoration struct {
	X, Y    uint16
	Content string
	Frames  int // remaining frames until removed
}

func (r *DefaultRenderer) Init() error {
	fmt.Printf("%s%s%s",
		"\033[?1049h", // Enable alternate buffer
		"\033[?25l",   // Make the cursor invisible
		"\033[J",      // Clear the screen
	)
	return nil
}

func (r *DefaultRenderer) Deinit() error {
	fmt.Printf("%s%s",
		"\033[?1049l", // Disable alternate buffer
		"\033[?25h",   // Make the cursor visible
	)
	return nil
}

func (r *DefaultRenderer) AddDecoration(col, row uint16, content string, frames int) {
	r.decorations = append(r.decorations, &decoration{
		X:       col,
		Y:       row,
		Content: content,
		Frames:  frames,
	})
	r.Fill(row, col, content)
}

func (r *DefaultRenderer) tickDecorations() {
	nd := []*decoration{}
	for _, d := range r.decorations {
		if d.Frames == 0 {
			r.Fill(d.Y, d.X, " ")
			continue
		}
		nd = append(nd, d)
		d.Frames--
	}
	r.decorations = nd
}

func (r *DefaultRenderer) RenderLoop(delay time.Duration, render func(startTime time.Time, duration time.Duration) bool) {
	cont := true
	startTime := time.Now().Add(delay)
	for cont {
		now := time.Now()
		duration := now.Sub(startTime)
		deadline := now.Add(*config.FramePeriod)

		cont = render(startTime, duration)

		r.tickDecorations()
		r.flush()

		remainingTime := deadline.Sub(time.Now())
		time.Sleep(remainingTime)
	}
}

func (r *DefaultRenderer) Fill(row, column uint16, message string) {
	r.buffer += fmt.Sprintf("\033[%d;%dH%v", row, column, message)
}

func (r *DefaultRenderer) FillColor(row, column uint16, c color.RGBA, message string) {
	r.buffer += fmt.Sprintf("\033[%d;%dH\033[38;2;%v;%v;%vm%v\033[0m", row, column, c.R, c.G, c.B, message)
}

func (r *DefaultRenderer) flush() {
	fmt.Print(r.buffer)
	r.buffer = ""
}
