package main

import (
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

const FailPenaltySeconds = 3
const FailPenaltyDuration = time.Second * FailPenaltySeconds
const FastErrorHighlightDuration = time.Millisecond * 333

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

	if s.Repeat {
		rep := "Repeating phrase"
		tbPrint(w-len(rep)-1, 1, termbox.ColorDefault, termbox.ColorDefault, rep)
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
