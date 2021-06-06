package game

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Judgement struct {
	Time  time.Duration
	Color rl.Color
	Name  string
}
