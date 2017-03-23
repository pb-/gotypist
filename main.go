package main

import (
	"bufio"
	"encoding/json"
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
	StartedAt time.Time
	FailedAt  time.Time
	Errors    int
}

type Phrase struct {
	Text   string
	Input  string
	Rounds [3]Round
	Mode   Mode
}

type State struct {
	Seed     int64
	Words    []string
	Timeouts map[time.Time]bool
	Phrase   Phrase
	Exiting  bool
}

type Statistics struct {
	Text       string    `json:"text"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Errors     int       `json:"errors"`
	Mode       Mode      `json:"mode"`
	Seconds    float64   `json:"seconds"`
	CPS        float64   `json:"cps"`
	WPM        float64   `json:"wpm"`
}

func NewPhrase(text string) *Phrase {
	return &Phrase{
		Text: text,
	}
}

func NewState(seed int64, words []string) *State {
	return &State{
		Timeouts: make(map[time.Time]bool),
		Seed:     seed,
		Words:    words,
		Phrase:   *NewPhrase(generateText(seed, words)),
	}
}

func (p *Phrase) CurrentRound() *Round {
	return &p.Rounds[p.Mode]
}

func (p *Phrase) ShowFail(t time.Time) bool {
	return p.Mode == ModeSlow && t.Sub(p.CurrentRound().FailedAt) < FailPenaltyDuration
}

func (p *Phrase) ErrorCountColor(t time.Time) termbox.Attribute {
	if p.Mode == ModeFast && t.Sub(p.CurrentRound().FailedAt) < FastErrorHighlightDuration {
		return termbox.ColorYellow | termbox.AttrBold
	}
	return termbox.ColorDefault
}

func (p *Phrase) IsErrorWith(ch rune) bool {
	input := p.Input + string(ch)
	return len(input) > len(p.Text) || input != p.Text[:len(input)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func nextSeed(seed int64) int64 {
	return rand.New(rand.NewSource(seed)).Int63()
}

func generateText(seed int64, words []string) string {
	rand := rand.New(rand.NewSource(seed))
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

	byteOffset, runeOffset := errorOffset(s.Phrase.Text, s.Phrase.Input)

	if s.Phrase.ShowFail(now) {
		left := min(int(s.Phrase.CurrentRound().FailedAt.Add(FailPenaltyDuration).Sub(now).Seconds()+1), FailPenaltySeconds)
		msg := fmt.Sprintf("FAIL! Let's do this again in %d...", left)
		tbPrint((w/2)-(len(msg)/2), h/2, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault, msg)
	} else {
		tbPrint((w/2)-(len(s.Phrase.Text)/2), h/2, termbox.ColorWhite, termbox.ColorDefault, s.Phrase.Text+string('⏎'))

		tbPrints((w/2)-(len(s.Phrase.Text)/2), h/2, termbox.ColorGreen, termbox.ColorDefault, s.Phrase.Input[:byteOffset])
		tbPrints((w/2)-(len(s.Phrase.Text)/2)+runeOffset, h/2, termbox.ColorBlack, termbox.ColorRed, s.Phrase.Input[byteOffset:])
	}

	mode := fmt.Sprintf("In %s mode", s.Phrase.Mode.Name())
	tbPrint((w/2)-(len(mode)/2), h/2-4, s.Phrase.Mode.Attr(), termbox.ColorDefault, mode)
	modeDesc := fmt.Sprintf("(%s!)", s.Phrase.Mode.Desc())
	tbPrint((w/2)-(len(modeDesc)/2), h/2-3, termbox.ColorDefault, termbox.ColorDefault, modeDesc)

	seconds, cps, wpm := computeStats(s.Phrase.Input[:byteOffset], s.Phrase.CurrentRound().StartedAt, now)
	stats := fmt.Sprintf("%3d errors  %4.1f s  %5.2f cps  %3d wpm", s.Phrase.CurrentRound().Errors, seconds, cps, int(wpm))
	tbPrint((w/2)-(len(stats)/2), h/2+4, termbox.ColorDefault, termbox.ColorDefault, stats)

	errors := fmt.Sprintf("%3d errors", s.Phrase.CurrentRound().Errors)
	tbPrint((w/2)-(len(stats)/2), h/2+4, s.Phrase.ErrorCountColor(now), termbox.ColorDefault, errors)

	tbPrint(1, h-3, termbox.ColorDefault, termbox.ColorDefault,
		"What's this fast, slow, medium thing?!")
	tbPrint(1, h-2, termbox.ColorDefault, termbox.ColorDefault,
		"http://steve-yegge.blogspot.com/2008/09/programmings-dirtiest-little-secret.html")
}

func reduce(s State, ev termbox.Event, now time.Time) State {
	if s.Phrase.ShowFail(now) {
		return s
	}

	if s.Phrase.CurrentRound().StartedAt.IsZero() {
		s.Phrase.CurrentRound().StartedAt = now
	}

	switch ev.Key {
	case termbox.KeyEsc, termbox.KeyCtrlC:
		s.Exiting = true
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(s.Phrase.Input) > 0 {
			_, l := utf8.DecodeLastRuneInString(s.Phrase.Input)
			s.Phrase.Input = s.Phrase.Input[:len(s.Phrase.Input)-l]
		}
	case termbox.KeyCtrlF:
		s.Seed = nextSeed(s.Seed)
		s.Phrase = *NewPhrase(generateText(s.Seed, s.Words))
	case termbox.KeyEnter:
		if s.Phrase.Input == s.Phrase.Text {
			if s.Phrase.Mode != ModeNormal {
				s.Phrase.Mode++
				s.Phrase.Input = ""
			} else {
				s.Seed = nextSeed(s.Seed)
				s.Phrase = *NewPhrase(generateText(s.Seed, s.Words))
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
			if s.Phrase.IsErrorWith(ch) {
				s.Phrase.CurrentRound().Errors++
				s.Phrase.CurrentRound().FailedAt = now
				if s.Phrase.Mode == ModeSlow {
					s.Phrase.Input = ""
					for t := time.Duration(1); t <= FailPenaltySeconds; t++ {
						s.Timeouts[now.Add(time.Second*t)] = true
					}
				} else if s.Phrase.Mode == ModeNormal {
					s.Phrase.Input += string(ch)
				} else if s.Phrase.Mode == ModeFast {
					s.Timeouts[now.Add(FastErrorHighlightDuration)] = true
				}
			} else {
				s.Phrase.Input += string(ch)
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

func must(errArg int) func(...interface{}) {
	return func(retval ...interface{}) {
		if err := retval[errArg]; err != nil {
			panic(err)
		}
	}
}

func logStatistics(phrase *Phrase, ev termbox.Event, now time.Time) {
	if ev.Key != termbox.KeyEnter || phrase.Input != phrase.Text {
		return
	}

	seconds, cps, wpm := computeStats(
		phrase.Text, phrase.CurrentRound().StartedAt, now)
	stats := Statistics{
		Text:       phrase.Text,
		StartedAt:  phrase.CurrentRound().StartedAt,
		FinishedAt: now,
		Errors:     phrase.CurrentRound().Errors,
		Mode:       phrase.Mode,
		Seconds:    seconds,
		CPS:        cps,
		WPM:        wpm,
	}

	f, err := os.OpenFile(os.Getenv("HOME")+"/.gotype.stats",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}

	must(1)(f.Write(data))
	must(1)(f.Write([]byte("\n")))
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

	state := *NewState(time.Now().UnixNano(), getWords("/usr/share/dict/words"))
	timers := make(map[time.Time]bool)

	render(state, time.Now())
	for !state.Exiting {
		ev := termbox.PollEvent()
		now := time.Now()

		switch ev.Type {
		case termbox.EventKey:
			logStatistics(&state.Phrase, ev, now)
			state = reduce(state, ev, now)
		case termbox.EventError:
			panic(ev.Err)
		case termbox.EventInterrupt:
		}

		render(state, now)
		timers = manageTimers(timers, state.Timeouts, now, termbox.Interrupt)
	}

}
