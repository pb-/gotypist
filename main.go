package main

import (
	"os"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	if len(os.Args) == 2 && os.Args[1] == "-d" {
		runDemo()
	} else {
		loop(os.Args, env())
	}
}

func loop(args []string, env map[string]string) {
	state, cmds := Init(args, env)
	state = runCommands(state, cmds)

	for {
		render(state, time.Now())
		state = reduceMessages(state, waitForEvent(), time.Now())
	}
}

func runCommands(state State, commands []Command) State {
	for _, command := range commands {
		state = reduceMessages(state, RunCommand(command), time.Now())
	}

	return state
}

func reduceMessages(state State, messages []Message, now time.Time) State {
	for _, message := range messages {
		newState, commands := reduce(state, message, time.Now())
		state = runCommands(newState, commands)
	}

	return state
}

func waitForEvent() []Message {
	ev := termbox.PollEvent()
	switch ev.Type {
	case termbox.EventKey:
		return []Message{ev}
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventInterrupt:
	}

	return []Message{}
}

func env() map[string]string {
	vars := map[string]string{}

	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		vars[pair[0]] = pair[1]
	}

	return vars
}
