package main

import (
	"math/rand"
	"strings"
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
	Seed           int64
	Words          []string
	StaticPhrase   bool
	Timeouts       map[time.Time]bool
	Phrase         Phrase
	Repeat         bool
	Exiting        bool
	RageQuit       bool
	Score          float64
	LastScore      float64
	LastScoreUntil time.Time
}

func reduce(s State, ev termbox.Event, now time.Time) State {
	if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
		s.Exiting = true
		if s.Phrase.ShowFail(now) {
			s.RageQuit = true
		}
	}
	if s.Phrase.ShowFail(now) {
		return s
	}

	if s.Phrase.CurrentRound().StartedAt.IsZero() {
		s.Phrase.CurrentRound().StartedAt = now
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(s.Phrase.Input) > 0 {
			_, l := utf8.DecodeLastRuneInString(s.Phrase.Input)
			s.Phrase.Input = s.Phrase.Input[:len(s.Phrase.Input)-l]
		}
	case termbox.KeyCtrlF:
		s = resetPhrase(s, true)
	case termbox.KeyCtrlR:
		s.Repeat = !s.Repeat
	case termbox.KeyEnter, termbox.KeyCtrlJ:
		if s.Phrase.Input == s.Phrase.Text {
			s.Phrase.CurrentRound().FinishedAt = now
			if s.Phrase.Mode != ModeNormal {
				s.Phrase.Mode++
				s.Phrase.Input = ""
			} else {
				s.LastScoreUntil = now.Add(ScoreHighlightDuration)
				s.Timeouts[s.LastScoreUntil] = true
				score := mustComputeScore(s.Phrase)
				s.LastScore = score
				s.Score += score
				s = resetPhrase(s, false)
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
			exp := s.Phrase.expected()
			if ch != exp {
				if exp != 0 {
					s.Phrase.CurrentRound().Typos = append(
						s.Phrase.CurrentRound().Typos, Typo{
							Expected: string(exp),
							Actual:   string(ch),
						})
				}
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

func resetPhrase(state State, forceNext bool) State {
	if state.StaticPhrase {
		state.Phrase = *NewPhrase(strings.Join(state.Words, " "))
	} else {
		if !state.Repeat || forceNext {
			state.Seed = nextSeed(state.Seed)
		}
		state.Phrase = *NewPhrase(generateText(state.Seed, state.Words))
	}
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

	return weightedScore(
		scores[ModeFast],
		scores[ModeSlow],
		scores[ModeNormal])
}

func NewState(seed int64, words []string, staticPhrase bool) *State {
	s := resetPhrase(State{
		Timeouts:     make(map[time.Time]bool),
		Seed:         seed,
		Words:        words,
		StaticPhrase: staticPhrase,
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

func nextSeed(seed int64) int64 {
	return rand.New(rand.NewSource(seed)).Int63()
}
