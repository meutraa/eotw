package render

import (
	"image/color"
	"time"
)

type Renderer interface {
	Init() error
	Deinit() error
	AddDecoration(col, row uint16, content string, frames int)
	RenderLoop(delay time.Duration, render func(startTime time.Time, duration time.Duration) bool)
	Fill(row, column uint16, message string)
	FillColor(row, column uint16, color color.RGBA, message string)
}
