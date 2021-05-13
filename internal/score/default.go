package score

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
	"git.lost.host/meutraa/eott/internal/game"
	_ "github.com/mattn/go-sqlite3"
)

type DefaultScorer struct {
	db *sql.DB
}

type InputsCompact struct {
	Index uint8
	Times []time.Duration
}

func compactInputs(inputs *[]game.Input) []InputsCompact {
	colCount := uint8(0)
	for _, i := range *inputs {
		if i.Index > colCount {
			colCount = i.Index + 1
		}
	}
	ins := make([]InputsCompact, colCount)
	for i := range ins {
		ins[i].Index = uint8(i)
	}
	for _, i := range *inputs {
		ins[i.Index].Index = i.Index // Repeated but it does not matter
		ins[i.Index].Times = append(ins[i.Index].Times, i.HitTime)
	}
	return ins
}

func uncompactInputs(inputs []InputsCompact) *[]game.Input {
	ins := []game.Input{}
	for _, i := range inputs {
		for _, t := range i.Times {
			ins = append(ins, game.Input{Index: i.Index, HitTime: t})
		}
	}
	return &ins
}

func (s *DefaultScorer) Init() error {
	db, err := sql.Open("sqlite3", "./scores.db")
	if err != nil {
		return err
	}

	initStatement := `
	create table if not exists scores 
	  (
		  id integer not null primary key, 
		  sum text,
		  rate integer,
		  inputs bytearray
	  );
	`
	_, err = db.Exec(initStatement)
	if nil != err {
		return nil
	}

	s.db = db
	return nil
}

func (s *DefaultScorer) Deinit() {
	if nil != s.db {
		s.db.Close()
	}
}

func (s *DefaultScorer) hashChart(c *game.Chart) string {
	sum := sha256.Sum256([]byte(c.Difficulty.Section))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func (s *DefaultScorer) Save(c *game.Chart, inputs *[]game.Input, rate uint16) {
	data, err := json.Marshal(compactInputs(inputs))
	if nil != err {
		log.Println("unable to marshal notes", err)
		return
	}
	_, err = s.db.Exec("insert into scores(sum, rate, inputs) values(?, ?, ?)", s.hashChart(c), rate, data)
	if nil != err {
		log.Println("unable to save score")
		return
	}
}

func (s *DefaultScorer) Load(c *game.Chart) []History {
	histories := []History{}
	rows, err := s.db.Query("select sum, rate, inputs from scores where sum = ?", s.hashChart(c))
	if nil != err && err != sql.ErrNoRows {
		log.Println("unable to load scores", err)
		return histories
	}
	defer rows.Close()
	for rows.Next() {
		var sum string
		var notes []byte
		var rate uint16
		rows.Scan(&sum, &rate, &notes)
		var ns []InputsCompact
		err := json.Unmarshal(notes, &ns)
		if nil != err {
			log.Println("unable to unmarshal note history")
			continue
		}
		inputs := uncompactInputs(ns)
		histories = append(histories, History{
			Sum:    sum,
			Inputs: inputs,
			Rate:   rate,
		})
	}
	return histories
}

func abs(x time.Duration) time.Duration {
	if x < 0 {
		return -x
	}
	return x
}

func (s *DefaultScorer) Distance(rate uint16, expected, actual time.Duration) time.Duration {
	return time.Duration(expected*100/time.Duration(rate)) - actual
}

func (s *DefaultScorer) DistanceFloat(rate uint16, expected, actual time.Duration) time.Duration {
	return time.Duration(100*float64(expected)/float64(rate)) - actual
}

func (s *DefaultScorer) ApplyHistoryToChart(ch *game.Chart, history *History) *game.Chart {
	nn := make([]*game.Note, len(ch.Notes))
	for i, n := range ch.Notes {
		nnn := *n
		nn[i] = &nnn
	}
	chart := game.Chart{
		Notes:      nn,
		NoteCount:  ch.NoteCount,
		MineCount:  ch.MineCount,
		Difficulty: ch.Difficulty,
	}
	for _, input := range *history.Inputs {
		s.ApplyInputToChart(&chart, &input, history.Rate)
	}
	return &chart
}

// Binary search for closest note, game.Chart should be in order
func (s *DefaultScorer) GetClosestNote(chart *game.Chart, input *game.Input, rate uint16) (note *game.Note, distance, abs time.Duration) {
	targets := make([]*game.Note, 0, len(chart.Notes))

	for _, note := range chart.Notes {
		// Reasons this note is not a valid hit target
		if note.HitTime != 0 || note.IsMine || note.Index != input.Index {
			continue
		}
		targets = append(targets, note)
	}

	if len(targets) == 0 {
		return
	}
	// TODO: Probably crashes if only one target

	note, distance, abs = s.searchClosest(targets, input, rate, 0, len(targets)-1)
	if abs < config.Judgements[len(config.Judgements)-2].Time {
		note.HitTime = input.HitTime
		return
	} else {
		return nil, 0, 0
	}
}

func (s *DefaultScorer) searchClosest(targets []*game.Note, input *game.Input, rate uint16, start, end int) (note *game.Note, distance, absDistance time.Duration) {
	mid := (start + end) / 2
	// It is impossible to hit the exact time, so do not check for equality

	// One option left
	if start == end {
		note = targets[start]
		distance = s.Distance(rate, targets[start].Time, input.HitTime)
		absDistance = abs(distance)
		return
	}

	// If only two options left
	if start == end-1 {
		note = targets[start]
		distance = s.Distance(rate, targets[start].Time, input.HitTime)
		absDistance = abs(distance)

		endDistanceDir := s.Distance(rate, targets[end].Time, input.HitTime)
		endDistance := abs(endDistanceDir)

		if absDistance > endDistance {
			note = targets[end]
			distance = endDistanceDir
			absDistance = endDistance
			return
		}
		return

	}

	midDistance := s.Distance(rate, targets[mid].Time, input.HitTime)
	// Hit early
	if midDistance > 0 {
		return s.searchClosest(targets, input, rate, start, mid)
	} else {
		return s.searchClosest(targets, input, rate, mid, end)
	}
}

func (s *DefaultScorer) ApplyInputToChart(chart *game.Chart, input *game.Input, rate uint16) (note *game.Note, distance, absDistance time.Duration) {
	return s.GetClosestNote(chart, input, rate)
	/*var closestNote *game.Note
	absDistance := time.Hour * 24
	distance := time.Hour * 24

	for _, note := range chart.Notes {
		if note.HitTime != 0 || note.IsMine {
			continue
		}
		if note.Index != input.Index {
			continue
		}
		dd := s.Distance(rate, note.Time, input.HitTime)
		d := abs(dd)
		if d < absDistance {
			distance = dd
			absDistance = d
			closestNote = note
		} else if nil != closestNote {
			// already found the closest, and this d is > md
			break
		}
	}

	if nil != closestNote && absDistance < config.Judgements[len(config.Judgements)-2].Time {
		closestNote.HitTime = input.HitTime
		onHit(closestNote, distance, absDistance)
	}*/
}

func (s *DefaultScorer) Score(chart *game.Chart, history *History) Score {
	var score Score
	ch := s.ApplyHistoryToChart(chart, history)
	for _, n := range ch.Notes {
		if n.HitTime == 0 {
			if !n.IsMine {
				score.MissCount++
			}
			continue
		}
		score.TotalError += abs(s.Distance(history.Rate, n.Time, n.HitTime))
	}
	return score
}
