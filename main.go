package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

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
	wordFile := flag.String("w", "/usr/share/dict/words", "path to word list")
	codeFile := flag.String("c", "", "path to code file")
	demo := flag.Bool("d", false, "demo mode for screenshot")
	numberProb := flag.Float64("n", 0, "mix in numbers with given probability (with -w)")
	flag.Parse()

	if *demo {
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

	seed := time.Now().UnixNano()
	var phraseFunc PhraseFunc
	if len(flag.Args()) > 0 {
		phraseFunc = StaticPhrase(strings.Join(flag.Args(), " "))
	} else {
		if *codeFile != "" {
			lines, err := loadCodeFile(*codeFile)
			if err != nil || len(lines) == 0 {
				phraseFunc = DefaultPhrase
			} else {
				seed = 0
				phraseFunc = SequentialLine(lines)
			}
		} else {
			dict, err := loadDictionary(*wordFile)
			if err != nil || len(dict) == 0 {
				phraseFunc = DefaultPhrase
			} else {
				phraseFunc = RandomPhrase(dict, 30, *numberProb)
			}
		}
	}

	state := *NewState(seed, phraseFunc)
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
