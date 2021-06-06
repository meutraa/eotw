package testdata

import (
	"encoding/json"

	"git.lost.host/meutraa/eotw/internal/game"
)

func GetChart() (*game.Chart, error) {
	var chart game.Chart
	if err := json.Unmarshal([]byte(data), &chart); nil != err {
		return nil, err
	}
	return &chart, nil
}
