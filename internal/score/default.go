package score

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"math"
	"time"

	"git.lost.host/meutraa/eott/internal/game"
	_ "github.com/mattn/go-sqlite3"
)

type DefaultScorer struct {
	db *sql.DB
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
		  notes text
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

func (s *DefaultScorer) Save(c *game.Chart, rate float64) {
	notes, err := json.Marshal(c.Notes)
	if nil != err {
		log.Println("unable to marshal notes", err)
		return
	}
	_, err = s.db.Exec("insert into scores(sum, rate, notes) values(?, ?, ?)", s.hashChart(c), rate, string(notes))
	if nil != err {
		log.Println("unable to save score")
		return
	}
}

func (s *DefaultScorer) Load(c *game.Chart) []History {
	histories := []History{}
	rows, err := s.db.Query("select sum, rate, notes from scores where sum = ?", s.hashChart(c))
	if nil != err && err != sql.ErrNoRows {
		log.Println("unable to load scores", err)
		return histories
	}
	defer rows.Close()
	for rows.Next() {
		var sum, notes string
		var rate float64
		rows.Scan(&sum, &rate, &notes)
		var ns []*game.Note
		err := json.Unmarshal([]byte(notes), &ns)
		if nil != err {
			log.Println("unable to unmarshal note history")
			continue
		}
		histories = append(histories, History{
			Sum:   sum,
			Notes: ns,
			Rate:  rate,
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

func (s *DefaultScorer) Score(history History) Score {
	var score Score
	var ms int64
	for _, n := range history.Notes {
		if n.HitTime == 0 {
			score.MissCount++
			continue
		}
		ms += abs(s.Distance(history.Rate, n, n.HitTime))
	}

	score.TotalError = time.Duration(int64(time.Millisecond) * ms)
	return score
}
