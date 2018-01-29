package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

func Init(args []string, env map[string]string) (State, []Command) {
	commandLine := flag.NewFlagSet(args[0], flag.ContinueOnError)

	commandLine.String("w", "/usr/share/dict/words", "load word list from `FILE`")
	commandLine.String("c", "", "load code from `FILE`")
	commandLine.Bool("d", false, "demo mode for screenshot")
	commandLine.Float64("n", 0, "mix in numbers with `PROBABILITY` (use with -w)")

	err := commandLine.Parse(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			buf := new(bytes.Buffer)
			commandLine.SetOutput(buf)
			commandLine.PrintDefaults()
			return State{}, []Command{Exit{Status: 1, GoodbyeMessage: buf.String()}}
		}
		return State{}, []Command{Exit{Status: 1, GoodbyeMessage: err.Error()}}
	}

	state := *NewState(time.Now().UnixNano(), StaticPhrase("somethings happenin in here"))
	home, _ := env["HOME"]
	state.Statsfile = home + "/.gotypist.stats"

	return state, []Command{
		ReadFile{
			Filename: state.Statsfile,
			Success:  func(data []byte) Message { return StatsData{Data: data} },
		},
		PeriodicInterrupt{250 * time.Millisecond},
	}
}

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

// needs porting
//	var phraseFunc PhraseFunc
//	if len(flag.Args()) > 0 {
//		phraseFunc = StaticPhrase(strings.Join(flag.Args(), " "))
//	} else {
//		if *codeFile != "" {
//			lines, err := loadCodeFile(*codeFile)
//			if err != nil || len(lines) == 0 {
//				phraseFunc = DefaultPhrase
//			} else {
//				phraseFunc = SequentialLine(lines)
//			}
//		} else {
//			dict, err := loadDictionary(*wordFile)
//			if err != nil || len(dict) == 0 {
//				phraseFunc = DefaultPhrase
//			} else {
//				phraseFunc = RandomPhrase(dict, 30, *numberProb)
//			}
//		}
//	}
