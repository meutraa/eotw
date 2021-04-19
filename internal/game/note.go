package game

type Note struct {
	Index   int // The chart column
	Row     int // The current row this note is rendered on, for clearing
	Denom   int // The beat length, as a denominator, 4 = 1/4 beat
	IsMine  bool
	Miss    bool  // Has the note scrolled past the bottom edge of the terminal
	HitTime int64 // When the note was hit
	Ms      int64 // The time the note should be hit
}
