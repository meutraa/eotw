package score

import (
	"time"

	"git.lost.host/meutraa/eotw/internal/game"
)

type Scorer interface {
	Init() error
	Deinit()

	// Save the state of this performance
	Save(chart *game.Chart, inputs *[]game.Input, rate uint16)

	// Load up previous state for the chart
	Load(chart *game.Chart) []History

	Score(chart *game.Chart, history *History) Score
	ApplyInputToChart(chart *game.Chart, input *game.Input, rate uint16) (note *game.Note, distance, absDistance time.Duration)

	Distance(rate uint16, expected, actual time.Duration) time.Duration
}

type History struct {
	Sum    string
	Inputs *[]game.Input
	Rate   uint16
}

type Score struct {
	MissCount  uint64
	TotalError time.Duration
}
