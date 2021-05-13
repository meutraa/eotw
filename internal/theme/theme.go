package theme

type Theme interface {
	RenderMine(column uint16, denom int) string
	RenderNote(column uint16, denom int) string
	RenderHitField(index uint8) string
}
