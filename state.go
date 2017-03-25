package main

import (
	"math/rand"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

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
	Repeat   bool
	Exiting  bool
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
	case termbox.KeyCtrlR:
		s.Repeat = !s.Repeat
	case termbox.KeyEnter:
		if s.Phrase.Input == s.Phrase.Text {
			if s.Phrase.Mode != ModeNormal {
				s.Phrase.Mode++
				s.Phrase.Input = ""
			} else {
				if s.Repeat {
					s.Phrase = *NewPhrase(s.Phrase.Text)
				} else {
					s.Seed = nextSeed(s.Seed)
					s.Phrase = *NewPhrase(generateText(s.Seed, s.Words))
				}
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

func nextSeed(seed int64) int64 {
	return rand.New(rand.NewSource(seed)).Int63()
}
