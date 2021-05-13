package score

import (
	"log"
	"testing"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
	"git.lost.host/meutraa/eott/internal/game"
	"git.lost.host/meutraa/eott/internal/testdata"
)

var hitTests = map[game.Input](*game.Note){
	{Index: 0, HitTime: time.Millisecond * 10155}: {Index: 0, Time: 10155169811},
	{Index: 0, HitTime: time.Millisecond * 42838}: {Index: 0, Time: 42843849056},
	{Index: 3, HitTime: time.Millisecond * 42990}: {Index: 3, Time: 42985358490},
	{}: nil,
}

func TestApplyInputToChart(t *testing.T) {
	config.Judgements = []game.Judgement{{Time: 180 * time.Millisecond}, {}}

	scorer := DefaultScorer{}
	chart, err := testdata.GetChart()
	if nil != err {
		log.Fatalln("unable to parse chart", err)
		return
	}
	for input, expected := range hitTests {
		note, _, _ := scorer.ApplyInputToChart(chart, &input, 100)
		if note == nil && expected == nil {
			continue
		}
		if note == nil || note.Time != expected.Time || note.Index != expected.Index {
			t.Log("Chart    ", chart.Notes[142], uint64(chart.Notes[142].Time))
			t.Log("Note    ", note)
			t.Log("Expected", expected)
			t.Fail()
		}
	}
}
