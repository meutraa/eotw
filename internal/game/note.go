package game

import (
	"time"
)

type Note struct {
	Index   uint8 // The chart column
	Denom   int   // The beat length, as a denominator, 4 = 1/4 beat
	IsMine  bool
	Time    time.Duration // The time the note should be hit
	TimeEnd time.Duration // The time the note should be unhit

	// This is state
	HitTime     time.Duration // When the note was hit
	ReleaseTime time.Duration // When the note was released
	MissTime    time.Duration // When the note was missed
}
