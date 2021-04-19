package config

import (
	"git.lost.host/meutraa/eott/internal/game"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	Directory           = kingpin.Arg("directory", "Song/chart directory").Required().ExistingDir()
	Rate                = kingpin.Flag("rate", "Playback seed").Default("1.0").Short('r').Float64()
	Offset              = kingpin.Flag("offset", "Global offset").Default("0ms").Short('o').Duration()
	Delay               = kingpin.Flag("delay", "Start delay").Default("1.5s").Short('d').Duration()
	ColumnSpacing       = kingpin.Flag("spacing", "Columns between keys").Default("6").Short('S').Uint()
	RefreshRate         = kingpin.Flag("refresh-rate", "Monitor refresh rate").Default("240.0").Short('R').Float()
	FramePeriod         = kingpin.Flag("frame-period", "Render frame period").Default("1ms").Short('p').Duration()
	scrollSpeedModifier = kingpin.Flag("scroll-speed", "Scroll speed, lower is faster").Default("3").Short('s').Uint()
	keys                = kingpin.Flag("keys", "Keys").Default("_-mp").Short('k').String()
	BarRow              = kingpin.Flag("bar-row", "Console row to render hit bar").Default("8").Uint()
	Keys                []rune
	ScrollSpeed         float64
	MissDistance        float64 // I count off the screen as missed
	Judgements          []game.Judgement
	NKey                = 4
)

var ()

func init() {
	kingpin.Version("0.2.0")
	kingpin.Parse()

	ScrollSpeed = float64(*scrollSpeedModifier) * 1000 / *RefreshRate
	Keys = []rune(*keys)
	MissDistance = float64(*BarRow) * ScrollSpeed
	Judgements = []game.Judgement{
		{Ms: 5, Name: "      \033[1;31mE\033[38;5;208mx\033[1;33ma\033[1;32mc\033[38;5;153mt\033[0m"},
		{Ms: 10, Name: " \033[1;35mRidiculous\033[0m"},
		{Ms: 20, Name: "  \033[38;5;153mMarvelous\033[0m"},
		{Ms: 40, Name: "      \033[1;36mGreat\033[0m"},
		{Ms: 60, Name: "       \033[1;32mGood\033[0m"},
		{Ms: MissDistance, Name: "       \033[1;31mOkay\033[0m"},
		{Ms: -1, Name: "       \033[1;31mMiss\033[0m"},
	}
}