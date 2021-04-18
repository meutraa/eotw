package game

import "git.lost.host/meutraa/eott/internal/graphics"

type Note struct {
	Index  int
	Row    int64
	Denom  int
	Hit    bool
	IsMine bool
	Miss   bool
	// A list of co-ords that need clearing when remaining tick == 1
	MissFlash          *map[string]graphics.Point
	MissFlashRemaining int
	HitTime            int64
	Ms                 int64
}
