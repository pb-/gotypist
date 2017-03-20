package main

import (
	"fmt"
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

var modes = []string{"fast", "slow", "normal"}
var modeDescs = []string{
	"type as fast as you can",
	"do not make any mistake",
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

type round struct {
	startedAt  time.Time
	finishedAt time.Time
	failedAt   time.Time
	errors     int
}

type state struct {
	text   string
	input  string
	rounds [3]round
	mode   Mode
}

func (s state) ShowFail() bool {
	return time.Now().Sub(s.rounds[s.mode].failedAt).Seconds() < 3
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
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

func render(s state) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	w, h := termbox.Size()

	byteOffset, runeOffset := errorOffset(s.text, s.input)

	if s.ShowFail() {
		msg := "FAIL! Let's do this again..."
		tbPrint((w/2)-(len(msg)/2), h/2, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault, msg)
	} else {
		tbPrint((w/2)-(len(s.text)/2), h/2, termbox.ColorWhite, termbox.ColorDefault, s.text+string('⏎'))

		tbPrints((w/2)-(len(s.text)/2), h/2, termbox.ColorGreen, termbox.ColorDefault, s.input[:byteOffset])
		tbPrints((w/2)-(len(s.text)/2)+runeOffset, h/2, termbox.ColorBlack, termbox.ColorRed, s.input[byteOffset:])
	}

	mode := "In " + s.mode.Name() + " mode"
	tbPrint((w/2)-(len(mode)/2), h/2-4, s.mode.Attr(), termbox.ColorDefault, mode)
	modeDesc := "(" + s.mode.Desc() + "!)"
	tbPrint((w/2)-(len(modeDesc)/2), h/2-3, termbox.ColorDefault, termbox.ColorDefault, modeDesc)

	cps := 0.
	seconds := 0.
	wpm := 0.

	if !s.rounds[s.mode].startedAt.IsZero() {
		delta := time.Now().Sub(s.rounds[s.mode].startedAt)
		seconds = delta.Seconds()
		if seconds > 0. {
			runeCount := utf8.RuneCountInString(s.input[:byteOffset])
			cps = float64(runeCount) / seconds
			wordCount := len(strings.Split(s.input[:byteOffset], " "))
			wpm = float64(wordCount) * 60 / seconds
		}
	}

	stats := fmt.Sprintf("%3d errors, %4.1f s, %5.2f cps, %3d wpm", s.rounds[s.mode].errors, seconds, cps, int(wpm))
	tbPrint((w/2)-(len(stats)/2), h/2+4, termbox.ColorDefault, termbox.ColorDefault, stats)
}

func reduce(s state, ev termbox.Event) state {
	if s.ShowFail() {
		return s
	}

	if s.rounds[s.mode].startedAt.IsZero() {
		s.rounds[s.mode].startedAt = time.Now()
	}

	switch ev.Key {
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(s.input) > 0 {
			_, l := utf8.DecodeLastRuneInString(s.input)
			s.input = s.input[:len(s.input)-l]
		}
	case termbox.KeyEnter:
		if s.input == s.text {
			if s.mode != ModeNormal {
				s.mode++
				s.input = ""
			}
		}
	case termbox.KeySpace:
		s.input += " "
	default:
		if ev.Ch != 0 {
			s.input += string(ev.Ch)
			if len(s.input) > len(s.text) || s.input != s.text[:len(s.input)] {
				s.rounds[s.mode].errors++
			}
		}

	}

	if s.mode == ModeSlow && s.input != s.text[:len(s.input)] {
		s.input = ""
		s.rounds[s.mode].failedAt = time.Now()
	}

	return s
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	go func() {
		for _ = range time.Tick(time.Millisecond * 250) {
			termbox.Interrupt()
		}
	}()

	var state state
	state.text = "the quick brown fox jumps over the lazy dog"

	render(state)
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return
			}
			state = reduce(state, ev)
		case termbox.EventError:
			panic(ev.Err)
		case termbox.EventInterrupt:
		}

		render(state)
	}
}
