package theme

import (
	"fmt"

	"image/color"
)

type DefaultTheme struct {
}

func (t *DefaultTheme) RenderMine(column int, denom int) string {
	r, g, b := getNoteColor(1)
	return fmt.Sprintf("\033[38;2;%v;%v;%vm%v\033[0m", r, g, b, mineSym)
}

func (t *DefaultTheme) RenderNote(column int, denom int) string {
	r, g, b := getNoteColor(denom)
	return fmt.Sprintf("\033[38;2;%v;%v;%vm%v\033[0m", r, g, b, noteSym)
}

func (t *DefaultTheme) RenderHitField(column int) string {
	return barSym
}

const (
	mineSym = "⨯"
	noteSym = "⬤"
	barSym  = "-"
)

var (
	noteColors = map[int]color.RGBA{
		1:  {R: 236, G: 30, B: 0},    // 1/4 red
		2:  {R: 0, G: 118, B: 236},   // 1/8 blue
		3:  {R: 106, G: 0, B: 236},   // 1/12 purple
		4:  {R: 236, G: 195, B: 0},   // 1/16 yellow
		5:  {R: 106, G: 106, B: 106}, // 1/20 grey???
		6:  {R: 236, G: 0, B: 106},   // 1/24 pink
		8:  {R: 236, G: 128, B: 0},   // 1/32 orange
		12: {R: 173, G: 236, B: 236}, // 1/48 light blue
		16: {R: 0, G: 236, B: 128},   // 1/64 green
		24: {R: 106, G: 106, B: 106}, // 1/96 grey
		32: {R: 106, G: 106, B: 106}, // 1/128 grey
		48: {R: 110, G: 147, B: 89},  // 1/192 olive
		64: {R: 106, G: 106, B: 106}, // 1/256 grey
		-1: {R: 255, G: 255, B: 255}, // other white
	}
)

func getNoteColor(d int) (r, g, b uint8) {
	col, ok := noteColors[d]
	if !ok {
		col = noteColors[-1]
	}
	r, g, b = col.R, col.G, col.B
	return
}
