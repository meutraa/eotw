package game

type Difficulty struct {
	Name    string
	Msd     string
	Section string
	NKeys   uint8
}

var NKeyMap = map[string]uint8{
	"dance-single": 4,
	"dance-solo":   6,
	"dance-double": 8,
}
