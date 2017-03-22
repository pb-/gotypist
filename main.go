package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

type Mode int

const (
	ModeFast Mode = iota
	ModeSlow
	ModeNormal
)

const FailPenaltySeconds = 3
const FailPenaltyDuration = time.Second * FailPenaltySeconds
const FastErrorHighlightDuration = time.Millisecond * 333

var modes = []string{"fast", "slow", "normal"}
var modeDescs = []string{
	"type as fast as you can, ignore mistakes",
	"go slow, do not make any mistake",
	"type at normal speed, avoid mistakes",
}
var modeColor = []termbox.Attribute{
	termbox.ColorGreen | termbox.AttrBold,
	termbox.ColorMagenta | termbox.AttrBold,
	termbox.ColorYellow | termbox.AttrBold,
}

func (m Mode) Name() string {
	return modes[m]
}

func (m Mode) Desc() string {
	return modeDescs[m]
}

func (m Mode) Attr() termbox.Attribute {
	return modeColor[m]
}

type Round struct {
	StartedAt  time.Time
	FinishedAt time.Time
	FailedAt   time.Time
	Errors     int
}

type State struct {
	Rand     *rand.Rand
	Words    []string
	Timeouts map[time.Time]bool
	Text     string
	Input    string
	Rounds   [3]Round
	Mode     Mode
}

type Statistics struct {
	Text       string    `json:"text"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Errors     int       `json:"errors"`
	Mode       Mode      `json:"mode"`
}

func NewState(rand *rand.Rand, words []string) *State {
	return &State{
		Timeouts: make(map[time.Time]bool),
		Rand:     rand,
		Words:    words,
		Text:     generateText(rand, words),
	}
}

func (s *State) CurrentRound() *Round {
	return &s.Rounds[s.Mode]
}

func (s *State) ShowFail(t time.Time) bool {
	return s.Mode == ModeSlow && t.Sub(s.CurrentRound().FailedAt) < FailPenaltyDuration
}

func (s *State) HighlightError(t time.Time) bool {
	return s.Mode == ModeFast && t.Sub(s.CurrentRound().FailedAt) < FastErrorHighlightDuration
}

func (s *State) IsErrorWith(ch rune) bool {
	input := s.Input + string(ch)
	return len(input) > len(s.Text) || input != s.Text[:len(input)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateText(rand *rand.Rand, words []string) string {
	var w []string
	l := int(rand.Int31n(4) + 4)
	for len(w) < l {
		w = append(w, words[rand.Int31n(int32(len(words)))])
	}
	return strings.Join(w, " ")
}

func readWords(path string) ([]string, error) {
	var words []string

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open words: %s", err)
	}
	defer file.Close()

	pattern := regexp.MustCompile(`^[a-z]+$`)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return words, nil
			}
			return nil, fmt.Errorf("failed to read words: %s", err)
		}

		trimmed := strings.TrimSpace(line)
		if pattern.MatchString(trimmed) && len(trimmed) < 8 {
			words = append(words, trimmed)
		}
	}
}

func getWords(path string) []string {
	words, err := readWords(path)
	if err != nil || len(words) < 9 {
		return []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
	}
	return words
}

func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func tbPrints(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		if c == ' ' {
			c = '␣'
		}
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func errorOffset(text string, input string) (int, int) {
	runeOffset := 0
	for i, tr := range text {
		if i >= len(input) {
			return len(input), runeOffset
		}

		ir, _ := utf8.DecodeRuneInString(input[i:])
		if ir != tr {
			return i, runeOffset
		}

		runeOffset++
	}

	return min(len(input), len(text)), runeOffset
}

func computeStats(text string, start, end time.Time) (seconds, cps, wpm float64) {
	if !start.IsZero() {
		seconds = end.Sub(start).Seconds()
		if seconds > 0. {
			runeCount := utf8.RuneCountInString(text)
			cps = float64(runeCount) / seconds
			wordCount := len(strings.Split(text, " "))
			wpm = float64(wordCount) * 60 / seconds
		}
	}

	return seconds, cps, wpm
}

func render(s State, now time.Time) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	w, h := termbox.Size()

	byteOffset, runeOffset := errorOffset(s.Text, s.Input)

	if s.ShowFail(now) {
		left := min(int(s.CurrentRound().FailedAt.Add(FailPenaltyDuration).Sub(now).Seconds()+1), FailPenaltySeconds)
		msg := fmt.Sprintf("FAIL! Let's do this again in %d...", left)
		tbPrint((w/2)-(len(msg)/2), h/2, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault, msg)
	} else {
		tbPrint((w/2)-(len(s.Text)/2), h/2, termbox.ColorWhite, termbox.ColorDefault, s.Text+string('⏎'))

		tbPrints((w/2)-(len(s.Text)/2), h/2, termbox.ColorGreen, termbox.ColorDefault, s.Input[:byteOffset])
		tbPrints((w/2)-(len(s.Text)/2)+runeOffset, h/2, termbox.ColorBlack, termbox.ColorRed, s.Input[byteOffset:])
	}

	mode := fmt.Sprintf("In %s mode", s.Mode.Name())
	tbPrint((w/2)-(len(mode)/2), h/2-4, s.Mode.Attr(), termbox.ColorDefault, mode)
	modeDesc := fmt.Sprintf("(%s!)", s.Mode.Desc())
	tbPrint((w/2)-(len(modeDesc)/2), h/2-3, termbox.ColorDefault, termbox.ColorDefault, modeDesc)

	seconds, cps, wpm := computeStats(s.Input[:byteOffset], s.CurrentRound().StartedAt, now)
	stats := fmt.Sprintf("%3d errors  %4.1f s  %5.2f cps  %3d wpm", s.CurrentRound().Errors, seconds, cps, int(wpm))
	tbPrint((w/2)-(len(stats)/2), h/2+4, termbox.ColorDefault, termbox.ColorDefault, stats)

	var color termbox.Attribute
	if s.HighlightError(now) {
		color = termbox.ColorYellow | termbox.AttrBold
	} else {
		color = termbox.ColorDefault
	}
	errors := fmt.Sprintf("%3d errors", s.CurrentRound().Errors)
	tbPrint((w/2)-(len(stats)/2), h/2+4, color, termbox.ColorDefault, errors)

	tbPrint(1, h-3, termbox.ColorDefault, termbox.ColorDefault,
		"What's this fast, slow, medium thing?!")
	tbPrint(1, h-2, termbox.ColorDefault, termbox.ColorDefault,
		"http://steve-yegge.blogspot.com/2008/09/programmings-dirtiest-little-secret.html")
}

func reduce(s State, ev termbox.Event, now time.Time) State {
	if s.ShowFail(now) {
		return s
	}

	if s.CurrentRound().StartedAt.IsZero() {
		s.CurrentRound().StartedAt = now
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(s.Input) > 0 {
			_, l := utf8.DecodeLastRuneInString(s.Input)
			s.Input = s.Input[:len(s.Input)-l]
		}
	case termbox.KeyEnter:
		if s.Input == s.Text {
			s.CurrentRound().FinishedAt = now
			if s.Mode != ModeNormal {
				s.Mode++
				s.Input = ""
			} else {
				return *NewState(s.Rand, s.Words)
			}
		}
	default:
		var ch rune
		if ev.Key == termbox.KeySpace {
			ch = ' '
		} else {
			ch = ev.Ch
		}

		if ch != 0 {
			if s.IsErrorWith(ch) {
				s.CurrentRound().Errors++
				s.CurrentRound().FailedAt = now
				if s.Mode == ModeSlow {
					s.Input = ""
					for t := time.Duration(1); t <= FailPenaltySeconds; t++ {
						s.Timeouts[now.Add(time.Second*t)] = true
					}
				} else if s.Mode == ModeNormal {
					s.Input += string(ch)
				} else if s.Mode == ModeFast {
					s.Timeouts[now.Add(FastErrorHighlightDuration)] = true
				}
			} else {
				s.Input += string(ch)
			}
		}

	}

	for k := range s.Timeouts {
		if k.Before(now) {
			delete(s.Timeouts, k)
		}
	}

	return s
}

func manageTimers(timers, timeouts map[time.Time]bool, now time.Time, interruptFunc func()) map[time.Time]bool {
	// remove old timers
	for t := range timers {
		if _, ok := timeouts[t]; !ok {
			delete(timers, t)
		}
	}

	// set up new timers
	for t := range timeouts {
		if _, ok := timers[t]; !ok {
			timers[t] = true
			time.AfterFunc(t.Sub(now), interruptFunc)
		}
	}

	return timers
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	go func() {
		for range time.Tick(time.Millisecond * 250) {
			termbox.Interrupt()
		}
	}()

	state := *NewState(rand.New(rand.NewSource(time.Now().UnixNano())), getWords("/usr/share/dict/words"))
	timers := make(map[time.Time]bool)

	render(state, time.Now())
	for {
		ev := termbox.PollEvent()
		now := time.Now()

		switch ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				return
			}
			state = reduce(state, ev, now)
		case termbox.EventError:
			panic(ev.Err)
		case termbox.EventInterrupt:
		}

		render(state, now)
		timers = manageTimers(timers, state.Timeouts, now, termbox.Interrupt)
	}

}
