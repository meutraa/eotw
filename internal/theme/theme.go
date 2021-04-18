package theme

type Theme interface {
	RenderMine(column int, denom int) string
	RenderNote(column int, denom int) string
	RenderHitField(column int) string
}
