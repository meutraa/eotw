package theme

import (
	"fmt"

	"git.lost.host/meutraa/eott/internal/graphics"
)

type DefaultTheme struct {
}

func (t *DefaultTheme) RenderMine(column int, denom int) string {
	color := getNoteColor(1)
	return fmt.Sprintf("\033[38;2;%v;%v;%vm%v\033[0m", color.R, color.G, color.B, mineSym)
}

func (t *DefaultTheme) RenderNote(column int, denom int) string {
	color := getNoteColor(denom)
	return fmt.Sprintf("\033[38;2;%v;%v;%vm%v\033[0m", color.R, color.G, color.B, syms[column])
}

func (t *DefaultTheme) RenderHitField(column int) string {
	return barSyms[column]
}

const (
	mineSym = "⨯"
)

var (
	syms       = [...]string{"⬤", "⬤", "⬤", "⬤"}
	barSyms    = [...]string{"-", "-", "-", "-"}
	noteColors = map[int]graphics.Color{
		1:  {236, 30, 0},    // 1/4 red
		2:  {0, 118, 236},   // 1/8 blue
		3:  {106, 0, 236},   // 1/12 purple
		4:  {236, 195, 0},   // 1/16 yellow
		5:  {106, 106, 106}, // 1/20 grey???
		6:  {236, 0, 106},   // 1/24 pink
		8:  {236, 128, 0},   // 1/32 orange
		12: {173, 236, 236}, // 1/48 light blue
		16: {0, 236, 128},   // 1/64 green
		24: {106, 106, 106}, // 1/96 grey
		32: {106, 106, 106}, // 1/128 grey
		48: {110, 147, 89},  // 1/192 olive
		64: {106, 106, 106}, // 1/256 grey
		-1: {255, 255, 255}, // other white
	}
)

func getNoteColor(d int) graphics.Color {
	col, ok := noteColors[d]
	if !ok {
		return noteColors[-1]
	}
	return col
}
