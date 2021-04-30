package score

import (
	"time"

	"git.lost.host/meutraa/eott/internal/game"
)

type Scorer interface {
	Init() error
	Deinit()

	// Save the state of this performance
	Save(chart *game.Chart, inputs *[]game.Input, rate float64)

	// Load up previous state for the chart
	Load(chart *game.Chart) []History

	Score(chart *game.Chart, history *History) Score
	ApplyInputToChart(chart *game.Chart, input *game.Input, rate float64, onHit func(note *game.Note, distance, absDistance time.Duration))

	Distance(rate float64, n *game.Note, hitTime time.Duration) time.Duration
}

type History struct {
	Sum    string
	Inputs *[]game.Input
	Rate   float64
}

type Score struct {
	MissCount  uint64
	TotalError time.Duration
}
