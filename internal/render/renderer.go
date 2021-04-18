package render

import (
	"time"

	"git.lost.host/meutraa/eott/internal/graphics"
)

type Renderer interface {
	Init()
	Deinit()
	RenderLoop(delay time.Duration, render func(now, deadline time.Time, duration time.Duration) bool)
	Fill(row, column int64, message string)
	FillColor(row, column int64, color graphics.Color, message string)
	Flush()
}
