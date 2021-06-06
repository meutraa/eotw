package main

import (
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"git.lost.host/meutraa/eotw/internal/config"
	"git.lost.host/meutraa/eotw/internal/game"
)

func main() {
	config.Init()
	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

func getColumn(nKeys uint8, mc int32, index uint8) int32 {
	// 4 => 2
	mid := nKeys >> 1
	if index < mid {
		return mc - *config.ColumnSpacing*int32(nKeys-mid-index) + *config.ColumnSpacing>>1
	} else {
		return mc + *config.ColumnSpacing*int32(index-mid) + *config.ColumnSpacing>>1
	}
}

func judge(d time.Duration) (int, *game.Judgement) {
	for i, j := range config.Judgements {
		if d < j.Time {
			return i, &j
		}
	}
	// This should never happen, since a check for d < missDistance is made
	return -1, nil
}

func run() error {
	flags := rl.FlagVsyncHint | rl.FlagMsaa4xHint | rl.FlagWindowResizable
	rl.SetConfigFlags(byte(flags))

	rl.InitWindow(1080, 1360, "eotw")
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	rl.SetTargetFPS(int32(*config.RefreshRate))

	program := Program{}
	for !rl.IsWindowReady() {
		log.Println("Waiting on window")
		time.Sleep(time.Millisecond * 5)
	}

	if err := program.Init(); nil != err {
		return err
	}

	music := rl.LoadMusicStream(program.audioFile)
	music.Looping = false
	defer rl.UnloadMusicStream(music)
	program.music = &music
	program.musicLength = rl.GetMusicTimeLength(music)

	rl.SetMusicPitch(music, float32(*config.Rate)/100)

	go func() {
		time.Sleep(*config.Delay + *config.Offset)
		rl.PlayMusicStream(music)
	}()

	program.startTime = time.Now().Add(*config.Delay)

	for !rl.WindowShouldClose() {
		rl.UpdateMusicStream(music)
		if rl.IsWindowResized() {
			program.Resize()
		}
		duration := time.Since(program.startTime)

		program.Update(duration)
		program.Render(duration)
	}

	program.Scorer.Save(&program.chart, &program.inputs, *config.Rate)
	return nil
}
