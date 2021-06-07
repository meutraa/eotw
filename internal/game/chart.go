package game

type Chart struct {
	Notes      []*Note
	Measures   []*Measure
	NoteCounts []int64
	HoldCount  int64
	MineCount  int64
	Difficulty Difficulty

	// This is for rendering optimization
	NoteCountsAsStrings []string

	// This is state
	activeNotes    []*Note
	startNoteIndex int
	endNoteIndex   int

	activeMeasures    []*Measure
	startMeasureIndex int
	endMeasureIndex   int
}

func (c *Chart) Active() ([]*Note, int, int) {
	return c.activeNotes, c.startNoteIndex, c.endNoteIndex
}

func (c *Chart) SetActive(start int, end int) {
	c.activeNotes = c.Notes[start:end]
	c.startNoteIndex = start
	c.endNoteIndex = end
}

func (c *Chart) ActiveMeasures() ([]*Measure, int, int) {
	return c.activeMeasures, c.startMeasureIndex, c.endMeasureIndex
}

func (c *Chart) SetActiveMeasures(start int, end int) {
	c.activeMeasures = c.Measures[start:end]
	c.startMeasureIndex = start
	c.endMeasureIndex = end
}
