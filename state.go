// only pure code in this file (no side effects)
package main

import (
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

type Typo struct {
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type Round struct {
	StartedAt  time.Time
	FailedAt   time.Time
	FinishedAt time.Time
	Errors     int
	Typos      []Typo
}

type Phrase struct {
	Text   string
	Input  string
	Rounds [3]Round
	Mode   Mode
}

type State struct {
	Codelines        bool
	NumberProb       float64
	Seed             int64
	PhraseGenerator  PhraseFunc
	Phrase           Phrase
	HideFingers      bool
	Repeat           bool
	RageQuit         bool
	Statsfile        string
	Score            float64
	LastScore        float64
	LastScorePercent float64
	LastScoreUntil   time.Time
}

func reduce(s State, msg Message, now time.Time) (State, []Command) {
	switch m := msg.(type) {
	case error:
		return s, []Command{Exit{GoodbyeMessage: m.Error()}}
	case Datasource:
		return reduceDatasource(s, m.Data, now)
	case StatsData:
		s.Score = getTotalScore(m.Data)
		return s, Noop
	case termbox.Event:
		return reduceEvent(s, m, now)
	}

	return s, Noop
}

func reduceEvent(s State, ev termbox.Event, now time.Time) (State, []Command) {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		return s, []Command{Exit{GoodbyeMessage: banner(s, now)}}
	}
	if s.Phrase.ShowFail(now) {
		return s, Noop
	}

	if s.Phrase.CurrentRound().StartedAt.IsZero() {
		s.Phrase.CurrentRound().StartedAt = now
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		return reduceBackspace(s)
	case termbox.KeyCtrlF:
		s = resetPhrase(s, true)
	case termbox.KeyCtrlR:
		s.Repeat = !s.Repeat
	case termbox.KeyCtrlI:
		s.HideFingers = !s.HideFingers
	case termbox.KeyEnter, termbox.KeyCtrlJ:
		return reduceEnter(s, now)
	default:
		return reduceCharInput(s, ev, now)
	}

	return s, Noop
}

func reduceBackspace(s State) (State, []Command) {
	if len(s.Phrase.Input) == 0 {
		return s, Noop
	}

	_, l := utf8.DecodeLastRuneInString(s.Phrase.Input)
	s.Phrase.Input = s.Phrase.Input[:len(s.Phrase.Input)-l]
	return s, Noop
}

func reduceEnter(s State, now time.Time) (State, []Command) {
	if s.Phrase.Input != s.Phrase.Text {
		return s, Noop
	}

	logCmd := AppendFile{
		Filename: s.Statsfile,
		Data:     formatStats(&s.Phrase, now),
		Error:    PassError,
	}

	s.Phrase.CurrentRound().FinishedAt = now
	if s.Phrase.Mode != ModeNormal {
		s.Phrase.Mode++
		s.Phrase.Input = ""
		return s, []Command{logCmd}
	}

	s.LastScoreUntil = now.Add(ScoreHighlightDuration)
	score := mustComputeScore(s.Phrase)
	s.LastScore = score
	s.LastScorePercent = score / maxScore(s.Phrase.Text)
	s.Score += score
	s = resetPhrase(s, false)

	return s, []Command{logCmd, Interrupt{ScoreHighlightDuration}}
}

func reduceCharInput(s State, ev termbox.Event, now time.Time) (State, []Command) {
	var ch rune
	if ev.Key == termbox.KeySpace {
		ch = ' '
	} else {
		ch = ev.Ch
	}

	if ch == 0 {
		return s, Noop
	}

	exp := s.Phrase.expected()
	if ch == exp {
		s.Phrase.Input += string(ch)
		return s, Noop
	}

	if exp != 0 {
		s.Phrase.CurrentRound().Typos = append(
			s.Phrase.CurrentRound().Typos, Typo{
				Expected: string(exp),
				Actual:   string(ch),
			})
	}

	s.Phrase.CurrentRound().Errors++
	s.Phrase.CurrentRound().FailedAt = now

	if s.Phrase.Mode == ModeFast {
		return s, []Command{Interrupt{FastErrorHighlightDuration}}
	}

	if s.Phrase.Mode == ModeSlow {
		s.Phrase.Input = ""
		cmds := []Command{}
		for t := time.Duration(1); t <= FailPenaltySeconds; t++ {
			cmds = append(cmds, Interrupt{t * time.Second})
		}
		return s, cmds
	}

	// normal mode
	s.Phrase.Input += string(ch)
	return s, Noop
}

func reduceDatasource(state State, data []byte, now time.Time) (State, []Command) {
	var items []string
	var err error

	if state.Codelines {
		items, err = loadCodeFile(data)
	} else {
		items, err = loadDictionary(data)
	}

	if err != nil {
		return state, []Command{Exit{GoodbyeMessage: err.Error()}}
	} else if len(items) == 0 {
		return state, []Command{Exit{GoodbyeMessage: "datafile contains no usable data"}}
	}

	if state.Codelines {
		state.PhraseGenerator = SequentialLine(items)
	} else {
		state.Seed = now.UnixNano()
		state.PhraseGenerator = RandomPhrase(items, 30, state.NumberProb)
	}

	return resetPhrase(state, false), Noop
}

func resetPhrase(state State, forceNext bool) State {
	if !state.Repeat || forceNext {
		next, _ := state.PhraseGenerator(state.Seed)
		state.Seed = next
	}
	_, phrase := state.PhraseGenerator(state.Seed)
	state.Phrase = *NewPhrase(phrase)

	return state
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

func mustComputeScore(phrase Phrase) float64 {
	var scores [3]float64
	if len(scores) != len(phrase.Rounds) {
		panic("bad score computation")
	}

	for mode, round := range phrase.Rounds {
		var s float64
		time := round.FinishedAt.Sub(round.StartedAt)
		switch mode {
		case ModeFast.Num():
			s = speedScore(phrase.Text, time)
		case ModeSlow.Num():
			s = errorScore(phrase.Text, round.Errors)
		case ModeNormal.Num():
			s = score(phrase.Text, time, round.Errors)
		}

		scores[mode] = s
	}

	return finalScore(
		phrase.Text,
		scores[ModeFast],
		scores[ModeSlow],
		scores[ModeNormal])
}

func NewState(seed int64, phraseGenerator PhraseFunc) *State {
	s := resetPhrase(State{
		PhraseGenerator: phraseGenerator,
		Seed:            seed,
		HideFingers:     true,
	}, false)

	return &s
}

func NewPhrase(text string) *Phrase {
	return &Phrase{
		Text: text,
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
	} else if p.Mode == ModeFast {
		return termbox.ColorBlack
	}
	return termbox.ColorDefault
}

func (p *Phrase) expected() rune {
	if len(p.Input) >= len(p.Text) {
		return 0
	}

	expected, _ := utf8.DecodeRuneInString(p.Text[len(p.Input):])
	return expected
}

func banner(s State, t time.Time) string {
	if s.Phrase.ShowFail(t) {
		return `
 ____       _       ____   _____    ___    _   _   ___   _____
|  _ \     / \     / ___| | ____|  / _ \  | | | | |_ _| |_   _|
| |_) |   / _ \   | |  _  |  _|   | | | | | | | |  | |    | |
|  _ <   / ___ \  | |_| | | |___  | |_| | | |_| |  | |    | |
|_| \_\ /_/   \_\  \____| |_____|  \__\_\  \___/  |___|   |_|
`
	}

	return ""
}
