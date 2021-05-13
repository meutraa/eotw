package game

type Chart struct {
	Notes      []*Note
	NoteCount  int64
	MineCount  int64
	Difficulty Difficulty

	activeNotes    []*Note
	startNoteIndex int
	endNoteIndex   int
}

func (c *Chart) Active() ([]*Note, int, int) {
	return c.activeNotes, c.startNoteIndex, c.endNoteIndex
}

func (c *Chart) SetActive(start int, end int) {
	c.activeNotes = c.Notes[start:end]
	c.startNoteIndex = start
	c.endNoteIndex = end
}
