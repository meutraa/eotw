package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"golang.org/x/term"
)

const (
	globalOffset = 0.00
	frameRate    = 240
	missDistance = 220
	startDelay   = 3000 * time.Millisecond
	flashLength  = 60 // error flash length ms
	// If the note is 300ms away, base distance is 300 rows, this divides that
	speed         = 1000 / frameRate * 3
	bottomPadding = 8
	mineSym       = "⨯"
	framePeriod   = 1 * time.Millisecond // game loop/render deadline
)

var (
	syms       = [4]string{"⬤", "⬤", "⬤", "⬤"}
	keys       = [4]string{"_", "-", "m", "p"}
	barSyms    = [4]string{"◯", "◯", "◯", "◯"}
	judgements = []Judgement{
		{0, 6, "      \033[1;31mE\033[38;5;208mx\033[1;33ma\033[1;32mc\033[38;5;153mt\033[0m"},
		{1, 11, " \033[1;35mRidiculous\033[0m"},
		{2, 22, "  \033[38;5;153mMarvelous\033[0m"},
		{3, 45, "    \033[1;36mPerfect\033[0m"},
		{4, 90, "      \033[1;32mGreat\033[0m"},
		{5, 135, "       \033[1;33mGood\033[0m"},
		{6, missDistance, "       \033[1;31mMiss\033[0m"},
	}
	noteColors = map[int]string{
		1:  "\033[38;2;236;30;0m",    // 1/4 red
		2:  "\033[38;2;0;118;236m",   // 1/8 blue
		3:  "\033[38;2;106;0;236m",   // 1/12 purple
		4:  "\033[38;2;236;195;0m",   // 1/16 yellow
		5:  "\033[38;2;106;106;106m", // 1/20 grey???
		6:  "\033[38;2;236;0;106m",   // 1/24 pink
		8:  "\033[38;2;236;128;0m",   // 1/32 orange
		12: "\033[38;2;173;236;236m", // 1/48 light blue
		16: "\033[38;2;0;236;128m",   // 1/64 green
		24: "\033[38;2;106;106;106m", // 1/96 grey
		32: "\033[38;2;106;106;106m", // 1/128 grey
		48: "\033[38;2;110;147;89m",  // 1/192 olive
		64: "\033[38;2;106;106;106m", // 1/256 grey
	}
)

type Judgement struct {
	index int
	ms    float64
	name  string
}

func getColor(d int64) string {
	col, ok := noteColors[int(d)]
	if !ok {
		return "\033[0m"
	}
	return col
}

var song = flag.String("f", "", "path to dir containing mp3 & sm file")

func fill(x, y int64, c string) string {
	return fmt.Sprintf("\033[%d;%dH\033[0m%v\033[0m", y, x, c)
}

func fillColor(x, y int64, color, c string) string {
	return fmt.Sprintf("\033[%d;%dH%v%v\033[0m", y, x, color, c)
}

func side(line int, sc int64, message string) string {
	return fmt.Sprintf("\033[%v;%vH%v\033[0m", line, sc, message)
}

func main() {
	if err := run(); nil != err {
		log.Fatalln(err)
	}
}

type Note struct {
	col    int
	row    int64
	color  string
	hit    bool
	isMine bool
	miss   bool
	// A list of co-ords that need clearing when remaining tick == 1
	missFlash          *map[string]Point
	missFlashRemaining int
	hitTime            int64
	ms                 int64
}

type Point struct {
	c, r int64
}

func isRowInField(rc int64, row int64, hit bool) bool {
	return !hit && (row < rc && row > 0)
}

func (note *Note) step(now time.Time, currentDuration time.Duration, cis *[4]int64, rc int64) (str string, miss int) {
	// clear all existing renders
	col := cis[note.col]
	if isRowInField(rc, note.row, false) {
		str += fill(col, note.row, " ")
	}
	if note.missFlashRemaining > 1 {
		note.missFlashRemaining--
	} else if note.missFlashRemaining == 1 {
		note.missFlashRemaining--
		// clear flash
		for _, r := range *note.missFlash {
			str += fill(r.c, r.r, " ")
		}
	} // else 0 and does not exist anymore

	// Calculate the new row based on time
	nr := (rc - bottomPadding)
	distance := int64(math.Round(float64((note.ms - currentDuration.Milliseconds())) / float64(speed)))
	note.row = nr - distance

	// Is this row within the playing field?
	if !note.miss && note.row > rc && !note.hit && !note.isMine {
		miss = 1
		note.miss = true
		// flash the miss
		note.missFlashRemaining = flashLength
		cen := rc >> 1
		note.missFlash = &map[string]Point{
			"╭": {col - 1, cen - 1},
			"╮": {col + 1, cen - 1},
			"╰": {col - 1, cen},
			"╯": {col + 1, cen},
		}
		for c, r := range *note.missFlash {
			str += fillColor(r.c, r.r, "\033[1;31m", c)
		}
	} else if isRowInField(rc, note.row, note.hit) {
		if note.isMine {

			str += fillColor(col, note.row, "\033[1;31m", mineSym)
		} else {
			str += fillColor(col, note.row, note.color, syms[note.col])
		}
	}
	return
}

type Chart struct {
	notes     []*Note
	noteCount int64
	mineCount int64
}

type BPM struct {
	startingBeat float64
	value        float64
}

func getSecondsPerNote(rates []BPM, currentBeat float64, bpn float64) (float64, float64) {
	sel := float64(0.0)
	for _, bpm := range rates {
		if currentBeat >= bpm.startingBeat {
			sel = bpm.value
			// log.Println("set bpm to", bpm)
		} else {
			break
		}
	}
	secondsPerBeat := 60.0 / sel
	// log.Println("secondsPerBeat", secondsPerBeat)
	return sel, bpn * secondsPerBeat
}

type Difficulty struct {
	name    string
	msd     string
	section string
}

func parse(ch <-chan keyboard.KeyEvent, file string) (*Chart, error) {
	data, err := ioutil.ReadFile(file)
	if nil != err {
		return nil, err
	}

	str := strings.ReplaceAll(string(data), "\r", "")
	sections := strings.Split(str, "#NOTES:")
	meta := sections[0]
	notesSection := ""
	difficulties := []Difficulty{}
	// difs := []string{"Challenge", "Hard", "Beginner"}
	for _, section := range sections[1:] {
		// Todo difficulty selection
		lines := strings.SplitN(section, "\n", 7)
		if !strings.Contains(lines[1], "dance-single") {
			continue
		}
		difficulties = append(difficulties, Difficulty{
			name:    strings.TrimSpace(lines[3]),
			msd:     strings.TrimSpace(lines[4]),
			section: lines[6],
		})
	}

	for i, dif := range difficulties {
		fmt.Printf("%v\t%3v\t%v\n", i, dif.msd, dif.name)
	}
	key := <-ch
	index, err := strconv.ParseInt(string(key.Rune), 10, 64)
	if nil != err || index > int64(len(difficulties)-1) {
		return nil, err
	}

	notesSection = difficulties[index].section

	offset := 0.0
	bpms := []BPM{}

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
				bpms = append(bpms, BPM{
					startingBeat: sb,
					value:        bbbs,
				})
			}
		}
	}

	log.Println("Offset:", offset)
	log.Println("BPMs:", bpms)

	// Start time of first note
	t := offset + globalOffset
	currentBeat := float64(0.0)

	notes := []*Note{}
	mineCount := 0
	noteCount := 0

	blocks := strings.Split(notesSection, "\n,")
	for _, block := range blocks {
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
			bpm, secondsPerNote := getSecondsPerNote(bpms, currentBeat, beatsPerNote)

			createNote := func(col int, c string) *Note {
				log.Printf("(%v) %v/%v = %v%vth\033[0m", bpm, i, lineCount, getColor(denom), denom)
				if c == "M" {
					mineCount++
				} else {
					noteCount++
				}
				return &Note{
					col:    col,
					color:  getColor(denom),
					isMine: c == "M",
					ms:     int64(t * 1000),
				}
			}

			for i, c := range chs {
				if mapToNote(c) {
					notes = append(notes, createNote(i, string(c)))
				}
			}

			t += secondsPerNote
			currentBeat += beatsPerNote
		}
	}

	return &Chart{
		notes:     notes,
		noteCount: int64(noteCount),
		mineCount: int64(mineCount),
	}, nil
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

func mapToNote(ch byte) bool {
	t := string(ch)
	return t == "1" || t == "2" || t == "M"
}

// Returns an index of the judgement array
func judge(d float64) Judgement {
	for _, judgement := range judgements {
		if d < judgement.ms {
			return judgement
		}
	}
	return judgements[len(judgements)-1]
}

func run() error {
	flag.Parse()

	cc, rc, err := term.GetSize(int(os.Stdout.Fd()))
	if nil != err {
		return fmt.Errorf("unable to get terminal size: %w", err)
	}

	keyChannel, err := keyboard.GetKeys(128)
	if nil != err {
		return fmt.Errorf("unable to open keyboard: %w", err)
	}
	defer func() {
		if err := keyboard.Close(); nil != err {
			log.Println("unable to close keyboard %w", err)
		}
	}()

	var mp3File, ogg, chartFile string

	if err := filepath.Walk(*song, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".mp3") {
			mp3File = path
		} else if strings.HasSuffix(info.Name(), ".ogg") {
			ogg = path
		} else if strings.HasSuffix(info.Name(), ".sm") {
			chartFile = path
		}
		return nil
	}); nil != err {
		return fmt.Errorf("unable to walk song directory: %w", err)
	}

	if (mp3File == "" && ogg == "") || chartFile == "" {
		return errors.New("unable to find .sm and .mp3 file in given directory")
	}

	mc := int64(cc) >> 1
	space := int64(6)
	cis := &([4]int64{mc - space*3, mc - space, mc + space, mc + space*3})
	sideCol := mc - space*9

	chart, err := parse(keyChannel, chartFile)
	if nil != err {
		return err
	}

	audioFile := mp3File
	if ogg != "" {
		audioFile = ogg
	}
	log.Printf("Opening %v (%v)\n", audioFile, chartFile)
	f, err := os.Open(audioFile)
	if err != nil {
		return err
	}
	var streamer beep.StreamSeekCloser
	var format beep.Format
	if ogg != "" {
		streamer, format, err = vorbis.Decode(f)
	} else {
		streamer, format, err = mp3.Decode(f)
	}
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/60))

	// Clear the screen and hide the cursor
	fmt.Printf("\033[?1049h\033[?25l\033[H\033[J")
	defer func() {
		// Restore the terminal state
		fmt.Printf("\033[?1049l\033[?25h")
	}()

	go func() {
		time.Sleep(startDelay)
		speaker.Play(streamer)
	}()
	startTime := time.Now().Add(startDelay)
	score := 0.0
	counts := make([]int, len(judgements))

	// stdev
	sumOfDistance := 0.0
	mean := 0.0
	totalHits := 0.0
	stdev := 0.0

	for {
		now := time.Now()
		currentDuration := now.Sub(startTime)
		if currentDuration.Milliseconds()-5000 > chart.notes[len(chart.notes)-1].ms {
			break
		}

		deadline := now.Add(framePeriod)
		str := ""

		// get the key inputs that occured so far
		for i := 0; i < len(keyChannel); i++ {
			key := <-keyChannel
			if key.Key == keyboard.KeyEsc {
				goto end
			}
			var closestNote *Note
			distance := 10000000.0
			dirDistance := 1000000.0
			for _, note := range chart.notes {
				if (note.hit) ||
					(note.isMine) ||
					(note.col != 0 && string(key.Rune) == keys[0]) ||
					(note.col != 1 && string(key.Rune) == keys[1]) ||
					(note.col != 2 && string(key.Rune) == keys[2]) ||
					(note.col != 3 && string(key.Rune) == keys[3]) {
					continue
				}
				dd := float64(note.ms - currentDuration.Milliseconds())
				d := math.Abs(dd)
				// log.Printf("hiiiiiit %v %v = %v", note.ms, currentDuration.Milliseconds(), dd)
				if d < distance {
					dirDistance = dd
					distance = d
					closestNote = note
				} else if nil != closestNote {
					// already found the closest, and this d is > md
					break
				}
			}
			if nil != closestNote && distance < missDistance {
				score += distance
				totalHits += 1
				sumOfDistance += dirDistance
				judgement := judge(distance)
				counts[judgement.index]++
				closestNote.hitTime = currentDuration.Milliseconds()
				closestNote.hit = true
				if totalHits > 2 {
					stdev = 0.0
					mean = sumOfDistance / totalHits
					for _, n := range chart.notes {
						if !n.hit {
							continue
						}
						diff := float64(n.ms - n.hitTime)
						xi := diff - mean
						xi2 := xi * xi
						stdev += xi2
					}
					stdev /= (totalHits - 1)
					stdev = math.Sqrt(stdev)
				}
			}
		}

		// Render the hit bar
		for i, sym := range barSyms {
			str += fill(cis[i], int64(rc)-bottomPadding, sym)
		}

		// Render notes
		for _, line := range chart.notes {
			mulstr, miss := line.step(now, currentDuration, cis, int64(rc))
			str += mulstr
			counts[len(counts)-1] += miss
		}

		remainingTime := deadline.Sub(time.Now())
		renderTime := framePeriod.Nanoseconds() - remainingTime.Nanoseconds()

		str += side(2, sideCol, fmt.Sprintf("Render Time:  %5.0f µs", float64(renderTime)/1000.0))
		str += side(3, sideCol, fmt.Sprintf("  Idle Time:  %.1f%%", 100-100*float64(renderTime)/float64(framePeriod.Nanoseconds())))
		str += side(10, sideCol, fmt.Sprintf("   Error dt:  %6v", score))
		str += side(11, sideCol, fmt.Sprintf("      Stdev:  %6.2f", stdev))
		str += side(12, sideCol, fmt.Sprintf("       Mean:  %6.2f", mean))
		str += side(13, sideCol, fmt.Sprintf("      Total:  %6v", chart.noteCount))
		str += side(14, sideCol, fmt.Sprintf("      Mines:  %6v", chart.mineCount))
		for i, judgement := range judgements {
			str += side(18+i, sideCol, fmt.Sprintf("%v:  %6v", judgement.name, counts[i]))
		}
		fmt.Print(str)

		time.Sleep(remainingTime)
	}
end:
	_ = <-keyChannel
	return nil
}
