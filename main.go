package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	var wordFile = flag.String(
		"w", "/usr/share/dict/words", "path to word list")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	go func() {
		for range time.Tick(time.Millisecond * 250) {
			termbox.Interrupt()
		}
	}()

	var words []string
	if len(flag.Args()) > 0 {
		words = flag.Args()
	} else {
		words = getWords(*wordFile)
	}
	state := *NewState(time.Now().UnixNano(), words, len(flag.Args()) != 0)
	timers := make(map[time.Time]bool)

	render(state, time.Now())
	for !state.Exiting {
		ev := termbox.PollEvent()
		now := time.Now()

		switch ev.Type {
		case termbox.EventKey:
			logStats(&state.Phrase, ev.Key, now)
			state = reduce(state, ev, now)
		case termbox.EventError:
			panic(ev.Err)
		case termbox.EventInterrupt:
		}

		render(state, now)
		timers = manageTimers(timers, state.Timeouts, now, termbox.Interrupt)
	}

	termbox.Close()

	if state.RageQuit {
		fmt.Println("Ragequitting, eh?")
	}
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
