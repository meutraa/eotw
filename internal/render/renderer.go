package render

import (
	"image"
	"image/color"
	"time"
)

type Renderer interface {
	Init()
	Deinit()
	AddDecoration(col, row int, content string, frames int)
	RenderLoop(delay time.Duration, render func(startTime time.Time, duration time.Duration) bool)
	Fill(row, column int, message string)
	FillColor(row, column int, color color.RGBA, message string)
}

type Decoration struct {
	Point   image.Point
	Content string
	Frames  int // remaining frames until removed
}
