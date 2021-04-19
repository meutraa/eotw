package score

import (
	"time"

	"git.lost.host/meutraa/eott/internal/game"
)

type Scorer interface {
	Init() error
	Deinit()

	// Save the state of this performance
	Save(chart *game.Chart, rate float64)

	// Load up previous state for the chart
	Load(chart *game.Chart) []History

	Score(history History) Score

	Distance(rate float64, n *game.Note, hitTime int64) int64
}

type History struct {
	Sum   string
	Notes []*game.Note
	Rate  float64
}

type Score struct {
	MissCount  uint64
	TotalError time.Duration
}
