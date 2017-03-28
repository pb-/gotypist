package main

import (
	"encoding/json"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

type Statistics struct {
	Text       string    `json:"text"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Errors     int       `json:"errors"`
	Typos      []Typo    `json:"typos"`
	Mode       Mode      `json:"mode"`
	Seconds    float64   `json:"seconds"`
	CPS        float64   `json:"cps"`
	WPM        float64   `json:"wpm"`
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

func writeStats(data []byte) {
	filename := os.Getenv("HOME") + "/.gotypist.stats"
	if name := os.Getenv("STATSFILE"); name != "" {
		filename = name
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	must(1)(f.Write(data))
	must(1)(f.Write([]byte("\n")))
}

func formatStats(phrase *Phrase, now time.Time) []byte {
	typos := phrase.CurrentRound().Typos
	if typos == nil {
		typos = make([]Typo, 0)
	}

	seconds, cps, wpm := computeStats(
		phrase.Text, phrase.CurrentRound().StartedAt, now)
	stats := Statistics{
		Text:       phrase.Text,
		StartedAt:  phrase.CurrentRound().StartedAt,
		FinishedAt: now,
		Errors:     phrase.CurrentRound().Errors,
		Typos:      typos,
		Mode:       phrase.Mode,
		Seconds:    seconds,
		CPS:        cps,
		WPM:        wpm,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}

	return data
}

func logStats(phrase *Phrase, key termbox.Key, now time.Time) {
	if key != termbox.KeyEnter || phrase.Input != phrase.Text {
		return
	}

	writeStats(formatStats(phrase, now))
}
