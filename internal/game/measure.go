package game

import (
	"time"
)

type Measure struct {
	Denom int           // The beat length, as a denominator, 4 = 1/4 beat
	Time  time.Duration // The time the note should be hit
}
