package parser

import "git.lost.host/meutraa/eott/internal/game"

type Parser interface {
	Parse(file string) ([]*game.Chart, error)
}
