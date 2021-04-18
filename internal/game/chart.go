package game

type Chart struct {
	Notes      []*Note
	NoteCount  int64
	MineCount  int64
	Difficulty Difficulty
}
