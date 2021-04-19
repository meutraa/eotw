package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	Directory     = kingpin.Arg("directory", "Song/chart directory").Required().ExistingDir()
	Rate          = kingpin.Flag("rate", "Playback seed").Default("1.0").Short('r').Float64()
	Offset        = kingpin.Flag("offset", "Global offset in ms").Default("0").Short('o').Int64()
	Delay         = kingpin.Flag("delay", "Start delay in ms").Default("1500").Short('d').Int64()
	ColumnSpacing = kingpin.Flag("spacing", "Columns between keys").Default("6").Short('s').Int()
)

func init() {
	kingpin.Version("0.1.0")
	kingpin.Parse()
}
