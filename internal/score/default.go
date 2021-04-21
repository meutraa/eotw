package score

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"math"
	"time"

	"git.lost.host/meutraa/eott/internal/config"
	"git.lost.host/meutraa/eott/internal/game"
	_ "github.com/mattn/go-sqlite3"
)

type DefaultScorer struct {
	db *sql.DB
}

type InputsCompact struct {
	Index int
	Times []time.Duration
}

func compactInputs(inputs *[]game.Input) []InputsCompact {
	colCount := 0
	for _, i := range *inputs {
		if i.Index > colCount {
			colCount = i.Index + 1
		}
	}
	ins := make([]InputsCompact, colCount)
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
		  rate real,
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

func (s *DefaultScorer) Save(c *game.Chart, inputs *[]game.Input, rate float64) {
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
		var rate float64
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

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func (s *DefaultScorer) Distance(rate float64, n *game.Note, hitTime int64) int64 {
	return int64(math.Round(float64(n.Ms)/rate)) - hitTime
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
		s.ApplyInputToChart(&chart, &input, history.Rate, func(note *game.Note, distance, absDistance float64) {})
	}
	return &chart
}

func (s *DefaultScorer) ApplyInputToChart(chart *game.Chart, input *game.Input, rate float64, onHit func(note *game.Note, distance, absDistance float64)) {
	var closestNote *game.Note
	absDistance := 10000000.0
	distance := 1000000.0

	for _, note := range chart.Notes {
		if note.HitTime != 0 || note.IsMine {
			continue
		}
		if note.Index != input.Index {
			continue
		}
		dd := s.Distance(rate, note, input.HitTime.Milliseconds())
		d := math.Abs(float64(dd))
		if d < absDistance {
			distance = float64(dd)
			absDistance = d
			closestNote = note
		} else if nil != closestNote {
			// already found the closest, and this d is > md
			break
		}
	}

	if nil != closestNote && absDistance < config.MissDistance {
		closestNote.HitTime = input.HitTime.Milliseconds()
		onHit(closestNote, distance, absDistance)
	}
}

func (s *DefaultScorer) Score(chart *game.Chart, history *History) Score {
	var score Score
	var ms int64
	ch := s.ApplyHistoryToChart(chart, history)
	for _, n := range ch.Notes {
		if n.HitTime == 0 {
			score.MissCount++
			continue
		}
		ms += abs(s.Distance(history.Rate, n, n.HitTime))
	}

	score.TotalError = time.Duration(int64(time.Millisecond) * ms)
	return score
}
