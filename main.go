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
	msgs := []Message{}
	state, cmds := Init(args, env)

	render(state, time.Now())

	for {
		if len(cmds) > 0 {
			newMsgs := RunCommand(cmds[0])
			msgs = append(msgs, newMsgs...)
			cmds = cmds[1:]
		}

		msg, newMsgs, ok := selectMessage(msgs)
		msgs = newMsgs

		now := time.Now()
		if ok {
			newState, newCmds := reduce(state, msg, now)
			state = newState
			cmds = append(cmds, newCmds...)
		}

		render(state, now)
	}
}

func selectMessage(msgs []Message) (Message, []Message, bool) {
	if len(msgs) > 0 {
		return msgs[0], msgs[1:], true
	}

	ev := termbox.PollEvent()
	switch ev.Type {
	case termbox.EventKey:
		return ev, msgs, true
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventInterrupt:
	}

	return nil, msgs, false
}

func env() map[string]string {
	vars := map[string]string{}

	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		vars[pair[0]] = pair[1]
	}

	return vars
}
