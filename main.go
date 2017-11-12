package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	wordFile string
	demo     bool
	isCode   bool
)

func init() {
	flag.StringVar(&wordFile, "w", "/usr/share/dict/words", "path to word list")
	flag.BoolVar(&demo, "d", false, "demo mode for screenshot")
	flag.BoolVar(&isCode, "c", false, "Expect code in word list (i.e. use any line)")
}

func loop(state State) bool {
	timers := make(map[time.Time]bool)

	render(state, time.Now())
	for !state.Exiting {
		ev := termbox.PollEvent()
		now := time.Now()

		switch ev.Type {
		case termbox.EventKey:
			go logStats(state.Phrase, ev.Key, now)
			state = reduce(state, ev, now)
		case termbox.EventError:
			panic(ev.Err)
		case termbox.EventInterrupt:
		}

		render(state, now)
		timers = manageTimers(timers, state.Timeouts, now, termbox.Interrupt)
	}

	return state.RageQuit
}

func main() {
	flag.Parse()

	if demo {
		runDemo()
		return
	}

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
	staticPhrase := len(flag.Args()) > 0
	if staticPhrase {
		words = flag.Args()
	} else {
		words = getWords(wordFile)
	}

	state := *NewState(time.Now().UnixNano(), words, staticPhrase)
	state.Score = getTotalScore()
	rageQuit := loop(state)

	termbox.Close()

	if rageQuit {
		fmt.Println(banner)
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

const banner = `
 ____       _       ____   _____    ___    _   _   ___   _____
|  _ \     / \     / ___| | ____|  / _ \  | | | | |_ _| |_   _|
| |_) |   / _ \   | |  _  |  _|   | | | | | | | |  | |    | |
|  _ <   / ___ \  | |_| | | |___  | |_| | | |_| |  | |    | |
|_| \_\ /_/   \_\  \____| |_____|  \__\_\  \___/  |___|   |_|
`
