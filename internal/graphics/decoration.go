package graphics

type Decoration struct {
	Point   Point
	Content string
	Frames  int // remaining frames until removed
}
