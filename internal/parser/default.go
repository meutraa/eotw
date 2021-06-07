package parser

import (
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"
	"time"

	"git.lost.host/meutraa/eotw/internal/game"
)

type DefaultParser struct{}

func (p *DefaultParser) getSecondsPerNote(rates []game.BPM, currentBeat float64, bpn float64) (float64, float64) {
	sel := float64(0.0)
	for _, bpm := range rates {
		if currentBeat >= bpm.StartingBeat {
			sel = bpm.Value
			// log.Println("set bpm to", bpm)
		} else {
			break
		}
	}
	secondsPerBeat := 60.0 / sel
	// log.Println("secondsPerBeat", secondsPerBeat)
	return sel, bpn * secondsPerBeat
}

// 0 – No note
// 1 – Normal note
// 2 – Hold head
// 3 – Hold/Roll tail
// 4 – Roll head
// M – Mine (or other negative note)
// K – Automatic keysound
// L – Lift note
// F – Fake note

func (p *DefaultParser) mapToNote(ch byte) bool {
	t := string(ch)
	return t == "1" || t == "2" || t == "4" || t == "M"
}

func (p *DefaultParser) Parse(file string) ([]*game.Chart, error) {
	data, err := ioutil.ReadFile(file)
	if nil != err {
		return nil, err
	}

	str := strings.ReplaceAll(string(data), "\r", "")
	sections := strings.Split(str, "#NOTES:")
	meta := sections[0]
	difficulties := []game.Difficulty{}
	for _, section := range sections[1:] {
		lines := strings.SplitN(section, "\n", 7)
		chartType := strings.TrimSpace(lines[1])
		chartType = strings.TrimSuffix(chartType, ":")
		nKeys, ok := game.NKeyMap[chartType]
		if !ok {
			continue
		}
		difficulties = append(difficulties, game.Difficulty{
			Name:    strings.TrimSuffix(strings.TrimSpace(lines[3]), ":"),
			Msd:     strings.TrimSuffix(strings.TrimSpace(lines[4]), ":"),
			Section: lines[6],
			NKeys:   nKeys,
		})
	}

	offset := 0.0
	bpms := []game.BPM{}

	for _, mdl := range strings.Split(meta, "\n#") {
		mdl = strings.TrimSpace(mdl)
		if strings.HasPrefix(mdl, "OFFSET:") {
			mdl = strings.TrimPrefix(mdl, "OFFSET:")
			mdl = strings.TrimSuffix(mdl, ";")
			offs, err := strconv.ParseFloat(mdl, 64)
			if nil != err {
				return nil, err
			}
			offset = -offs
		} else if strings.HasPrefix(mdl, "BPMS:") {
			mdl = strings.TrimPrefix(mdl, "BPMS:")
			mdl = strings.ReplaceAll(mdl, "\n", "")
			bbs := strings.Split(strings.TrimSuffix(mdl, ";"), ",")
			for _, bpm := range bbs {
				as := strings.Split(bpm, "=")
				sb, err := strconv.ParseFloat(as[0], 64)
				if nil != err {
					return nil, err
				}
				bbbs, err := strconv.ParseFloat(as[1], 64)
				if nil != err {
					return nil, err
				}
				bpms = append(bpms, game.BPM{
					StartingBeat: sb,
					Value:        bbbs,
				})
			}
		}
	}

	charts := []*game.Chart{}
	for _, difficulty := range difficulties {
		// Start time of first note
		seconds := offset
		var currentBeat float64 = 0.0

		notes := []*game.Note{}
		mineCount := 0
		holdCount := 0
		noteCounts := make([]int64, difficulty.NKeys)

		blocks := strings.Split(difficulty.Section, "\n,")
		measureTimes := []*game.Measure{}

		for _, block := range blocks {
			measureTimes = append(measureTimes, &game.Measure{
				Denom: 1,
				Time:  time.Duration(seconds * 1000 * 1000 * 1000),
			})

			lines := []string{}
			bls := strings.Split(block, "\n")
			for _, l := range bls {
				if strings.HasPrefix(l, " ") || strings.Contains(l, "-") {
					continue
				}
				l = strings.TrimSpace(l)
				if len(l) > 3 {
					lines = append(lines, l)
				}
			}

			// Beat count is 4 per block
			lineCount := int64(len(lines))
			beatsPerNote := 4.0 / float64(lineCount) // 1/4, 1/8, 1/16, 1/24 etc

			// for each note line in a block
			for i, line := range lines {
				chs := []byte(line)
				r := big.NewRat(int64(i*4), lineCount)
				denom := r.Denom().Int64()
				if denom == 1 && i != 0 {
					measureTimes = append(measureTimes, &game.Measure{
						Denom: 4,
						Time:  time.Duration(seconds * 1000 * 1000 * 1000),
					})
				}
				if denom == 2 || denom == 4 {
					measureTimes = append(measureTimes, &game.Measure{
						Denom: 8,
						Time:  time.Duration(seconds * 1000 * 1000 * 1000),
					})
				}
				_, secondsPerNote := p.getSecondsPerNote(bpms, currentBeat, beatsPerNote)

				createNote := func(index uint8, c byte) *game.Note {
					// log.Printf("(%v) %v/%v = %v%vth\033[0m", bpm, i, lineCount, (denom), denom)
					if c == 'M' {
						mineCount++
					} else if c == '2' || c == '4' {
						holdCount++
					}
					return &game.Note{
						Index:  index,
						Denom:  int(denom),
						IsMine: c == 'M',
						Time:   time.Duration(seconds * 1000 * 1000 * 1000),
					}
				}

				hitCount := 0
				for i, c := range chs {
					// Positive hits at the same time
					if c == '1' || c == '2' || c == '4' {
						hitCount++
					}

					if p.mapToNote(c) {
						notes = append(notes, createNote(uint8(i), c))
					} else if c == '3' {
						// This is a release note of a previous head
						// Find the last note of type head in this column and
						// add this as the endtime to it
						// Loop through notes in reverse
						for j := len(notes) - 1; j >= 0; j-- {
							note := notes[j]
							if int(note.Index) != i {
								continue
							}

							// This will be the matching note
							note.TimeEnd = time.Duration(seconds * 1000 * 1000 * 1000)
							break
						}
					}
				}

				if hitCount > 0 {
					noteCounts[hitCount-1] += 1
				}

				seconds += secondsPerNote
				currentBeat += beatsPerNote
			}
		}

		noteCountsAsStrings := make([]string, difficulty.NKeys)
		for i, count := range noteCounts {
			noteCountsAsStrings[i] = strconv.FormatInt(count, 10)
		}

		charts = append(charts, &game.Chart{
			Notes:               notes,
			Measures:            measureTimes,
			NoteCounts:          noteCounts,
			NoteCountsAsStrings: noteCountsAsStrings,
			HoldCount:           int64(holdCount),
			MineCount:           int64(mineCount),
			Difficulty:          difficulty,
		})
	}

	return charts, nil
}
