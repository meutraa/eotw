package game

import "time"

type Input struct {
	Index   uint8 // game column index
	HitTime time.Duration
}
