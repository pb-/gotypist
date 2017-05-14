package main

import (
	"time"

	"github.com/nsf/termbox-go"
)

func runDemo() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	state := *NewState(
		0, []string{"correct", "horse", "battery", "staple"}, true)
	now := time.Now()

	state.Score = 8938
	state.LastScoreUntil = now.Add(1 * time.Second)
	state.LastScore = 201
	state.LastScorePercent = 0.87
	state.Phrase.Mode = ModeNormal
	state.Phrase.CurrentRound().Errors = 2
	state.Phrase.CurrentRound().StartedAt = now.Add(-1941 * time.Millisecond)
	state.Phrase.Input = "correct horse bta"
	state.HideFingers = false

	for {
		render(state, now)
		if ev := termbox.PollEvent(); ev.Type == termbox.EventKey {
			return
		}
	}
}
