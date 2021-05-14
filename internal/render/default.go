package render

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
	"golang.org/x/crypto/ssh/terminal"
)

type DefaultRenderer struct {
	buffer       strings.Builder
	restoreState *terminal.State
	decorations  []*decoration
}

type decoration struct {
	X, Y    uint16
	Content string
	Frames  int // remaining frames until removed
}

func (r *DefaultRenderer) Init() error {
	state, err := terminal.MakeRaw(int(os.Stdout.Fd()))
	if nil != err {
		return err
	}
	r.restoreState = state

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
	return terminal.Restore(int(os.Stdout.Fd()), r.restoreState)
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
	nd := make([]*decoration, 0, len(r.decorations))
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

func (r *DefaultRenderer) RenderLoop(
	delay time.Duration,
	render func(startTime time.Time, duration time.Duration) bool,
	endRender func(renderDuration time.Duration),
) {
	cont := true
	startTime := time.Now().Add(delay)
	for cont {
		now := time.Now()
		duration := now.Sub(startTime)
		deadline := now.Add(*config.FramePeriod)

		cont = render(startTime, duration)

		r.tickDecorations()
		r.flush()

		endRender(time.Now().Sub(now))
		remainingTime := deadline.Sub(time.Now())
		time.Sleep(remainingTime)
	}
}

func (r *DefaultRenderer) Fill(row, column uint16, message string) {
	r.buffer.WriteString("\033[")
	r.buffer.WriteString(strconv.FormatInt(int64(row), 10))
	r.buffer.WriteString(";")
	r.buffer.WriteString(strconv.FormatInt(int64(column), 10))
	r.buffer.WriteString("H")
	r.buffer.WriteString(message)
}

func (r *DefaultRenderer) FillColor(row, column uint16, c color.RGBA, message string) {
	r.buffer.WriteString("\033[")
	r.buffer.WriteString(strconv.FormatInt(int64(row), 10))
	r.buffer.WriteString(";")
	r.buffer.WriteString(strconv.FormatInt(int64(column), 10))
	r.buffer.WriteString("H\033[38;2;")
	r.buffer.WriteString(strconv.FormatInt(int64(c.R), 10))
	r.buffer.WriteString(";")
	r.buffer.WriteString(strconv.FormatInt(int64(c.G), 10))
	r.buffer.WriteString(";")
	r.buffer.WriteString(strconv.FormatInt(int64(c.B), 10))
	r.buffer.WriteString(message)
	r.buffer.WriteString("\033[0m")
}

func (r *DefaultRenderer) flush() {
	os.Stdout.Write([]byte(r.buffer.String()))
	r.buffer.Reset()
}
