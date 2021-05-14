package config

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"git.lost.host/meutraa/eott/internal/game"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	Directory           = kingpin.Arg("directory", "Song/chart directory").Required().ExistingDir()
	CpuProfile          = kingpin.Flag("profile", "Profile CPU").String()
	DebugUpdateRate     = kingpin.Flag("debug-update-rate", "Every n frames").Default("240").Uint16()
	Input               = kingpin.Flag("input", "Input device").Default("/dev/input/by-id/usb-OLKB_Planck-event-kbd").Short('i').ExistingFile()
	Rate                = kingpin.Flag("rate", "Playback % rate").Default("100").Short('r').Uint16()
	Offset              = kingpin.Flag("offset", "Global offset").Default("0ms").Short('o').Duration()
	Delay               = kingpin.Flag("delay", "Start delay").Default("1.5s").Short('d').Duration()
	ColumnSpacing       = kingpin.Flag("spacing", "Columns between keys").Default("12").Short('S').Uint16()
	RefreshRate         = kingpin.Flag("refresh-rate", "Monitor refresh rate").Default("240.0").Short('R').Float()
	FramePeriod         = kingpin.Flag("frame-period", "Render frame period").Default("1ms").Short('p').Duration()
	scrollSpeedModifier = kingpin.Flag("scroll-speed", "Scroll speed, lower is faster").Default("3").Short('s').Uint()
	keys4               = kingpin.Flag("keys-single", "Keys for 4k").Default("12,40,17,50").Short('k').String()
	keys6               = kingpin.Flag("keys-solo", "Keys for 6k").Default("23,18,24,20,31,46").String()
	keys8               = kingpin.Flag("keys-double", "Keys for 8k").Default("23,18,24,49,35,20,31,46").String()
	BarOffsetFromBottom = kingpin.Flag("bar-row", "Console row to render hit bar").Default("8").Uint16()
	BarSym              = kingpin.Flag("bar-decoration", "Decoration at the hitfield").Default("\033[2m\033[1D[ ]").String()
	NoteSym             = kingpin.Flag("note-symbol", "Restricted to 1 column").Default("⬤").String()
	MineSym             = kingpin.Flag("mine-symbol", "Restricted to 1 column").Default("⨯").String()

	Keys4      [4]uint16
	Keys6      [6]uint16
	Keys8      [8]uint16
	NsToRow    float64
	Judgements []game.Judgement
)

func Keys(nKeys uint8) []uint16 {
	switch nKeys {
	case 4:
		return Keys4[:]
	case 6:
		return Keys6[:]
	case 8:
		return Keys8[:]
	}
	return Keys4[:]
}

func KeyColumn(r uint16, nKeys uint8) (uint8, error) {
	for i, c := range Keys(nKeys) {
		if r == c {
			return uint8(i), nil
		}
	}
	return 0, errors.New("Key not mapped to index")
}

func Init() {
	kingpin.Version("0.2.0")
	kingpin.Parse()

	keys := strings.Split(*keys4, ",")
	for i, key := range keys {
		p, err := strconv.ParseUint(key, 10, 16)
		if nil != err {
			log.Fatalln(err)
		}
		Keys4[i] = uint16(p)
	}

	NsToRow = 1 / (float64(*scrollSpeedModifier) * 1000 / *RefreshRate * 1000000)
	Judgements = []game.Judgement{
		{Time: 11 * time.Millisecond, Name: "      \033[1;31mE\033[38;5;208mx\033[1;33ma\033[1;32mc\033[38;5;153mt\033[0m"},
		{Time: 22 * time.Millisecond, Name: "  \033[38;5;141mMarvelous\033[0m"},
		{Time: 45 * time.Millisecond, Name: "    \033[38;5;117mPerfect\033[0m"},
		{Time: 90 * time.Millisecond, Name: "      \033[38;5;155mGreat\033[0m"},
		{Time: 135 * time.Millisecond, Name: "       \033[38;5;214mGood\033[0m"},
		{Time: 180 * time.Millisecond, Name: "        \033[38;5;208mBoo\033[0m"},
		{Time: -1, Name: "       \033[38;5;160mMiss\033[0m"},
	}
}
