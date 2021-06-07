package config

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"git.lost.host/meutraa/eotw/internal/game"
	rl "github.com/gen2brain/raylib-go/raylib"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	Directory           = kingpin.Arg("directory", "Song/chart directory").Required().ExistingDir()
	Rate                = kingpin.Flag("rate", "Playback % rate").Default("100").Short('r').Uint16()
	Offset              = kingpin.Flag("offset", "Global offset").Default("0ms").Short('o').Duration()
	Delay               = kingpin.Flag("delay", "Start delay").Default("1.5s").Short('d').Duration()
	ColumnSpacing       = kingpin.Flag("spacing", "Columns between keys").Default("120").Short('S').Int32()
	RefreshRate         = kingpin.Flag("refresh-rate", "Monitor refresh rate").Default("240.0").Short('R').Float()
	NoteRadius          = kingpin.Flag("note-radius", "Radius of notes").Default("14").Float32()
	scrollSpeedModifier = kingpin.Flag("scroll-speed", "Scroll speed, lower is faster").Default("3").Short('s').Uint()
	keys4               = kingpin.Flag("keys-single", "Keys for 4k").Default("73,69,83,67").Short('k').String()
	keys6               = kingpin.Flag("keys-solo", "Keys for 6k").Default("23,18,24,20,31,46").String()
	keys8               = kingpin.Flag("keys-double", "Keys for 8k").Default("23,18,24,49,35,20,31,46").String()
	FontSize            = kingpin.Flag("font-size", "Font size").Default("24").Int32()
	BarOffsetFromBottom = kingpin.Flag("bar-row", "Pixels from bottom to render hit bar").Default("220").Int32()
	BarSym              = kingpin.Flag("bar-decoration", "Decoration at the hitfield").Default("\033[2m\033[1D[ ]").String()

	Keys4       [4]int32
	Keys6       [6]int32
	Keys8       [8]int32
	PixelsPerNs float64
	Judgements  []game.Judgement
)

func Keys(nKeys uint8) []int32 {
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

func KeyColumn(r int32, nKeys uint8) (uint8, error) {
	for i, c := range Keys(nKeys) {
		if r == c {
			return uint8(i), nil
		}
	}
	return 0, errors.New("key not mapped to index")
}

func Init() {
	kingpin.Version("0.2.0")
	kingpin.Parse()

	keys := strings.Split(*keys4, ",")
	for i, key := range keys {
		p, err := strconv.ParseInt(key, 10, 32)
		if nil != err {
			log.Fatalln(err)
		}
		Keys4[i] = int32(p)
	}

	PixelsPerNs = 1 / (float64(*scrollSpeedModifier) * 40 / *RefreshRate * 1000000)
	Judgements = []game.Judgement{
		{Time: 11 * time.Millisecond,
			Name:  "      Exact",
			Color: rl.NewColor(63, 0, 255, 255),
		},
		{Time: 22 * time.Millisecond,
			Name:  "  Marvelous",
			Color: rl.NewColor(175, 135, 255, 255),
		},
		{Time: 45 * time.Millisecond,
			Name:  "    Perfect",
			Color: rl.NewColor(135, 215, 255, 255),
		},
		{Time: 90 * time.Millisecond,
			Name:  "      Great",
			Color: rl.NewColor(175, 255, 95, 255),
		},
		{Time: 135 * time.Millisecond,
			Name:  "       Good",
			Color: rl.NewColor(255, 175, 0, 255),
		},
		{Time: 180 * time.Millisecond,
			Name:  "        Boo",
			Color: rl.NewColor(255, 135, 0, 255),
		},
		{Time: -1,
			Name:  "       Miss",
			Color: rl.NewColor(215, 0, 0, 255),
		},
	}
}
