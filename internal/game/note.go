package game

import "time"

type Note struct {
	Index   int // The chart column
	Denom   int // The beat length, as a denominator, 4 = 1/4 beat
	IsMine  bool
	Time    time.Duration // The time the note should be hit
	TimeEnd time.Duration // TDurationhe time the note should be unhit

	// This is state
	Row     int           // The current row this note is rendered on, for clearing
	HitTime time.Duration // When the note was hit
	Miss    bool          // Has the note scrolled past the bottom edge of the terminal
}
