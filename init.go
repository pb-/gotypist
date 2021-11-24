// only pure code in this file
package main

import (
	"bytes"
	"flag"
	"strings"
	"time"
)

func Init(args []string, env map[string]string) (State, []Command) {
	state := *NewState(0, DefaultPhrase)

	commandLine := flag.NewFlagSet(args[0], flag.ContinueOnError)
	datafile := commandLine.String("f", "", "load word list from `FILE`. \"-\" for stdin.")
	commandLine.BoolVar(&state.Codelines, "c", false, "treat -f FILE as lines of code")
	commandLine.Bool("d", false, "demo mode for screenshot")
	commandLine.Float64Var(&state.NumberProb, "n", 0, "mix in numbers with `PROBABILITY`")

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

	commands := []Command{}

	if len(commandLine.Args()) > 0 {
		state.PhraseGenerator = StaticPhrase(strings.Join(commandLine.Args(), " "))
		state = resetPhrase(state, false)
	} else if *datafile == "" {
		commands = append(commands, LoadBuiltinDictionary{})
	} else {
		commands = append(commands, ReadFile{
			Filename: *datafile,
			Success:  func(data []byte) Message { return Datasource{Data: data} },
			Error:    PassError,
		})
	}

	home, _ := env["HOME"]
	state.Statsfile = home + "/.gotypist.stats"

	return state, append(commands,
		ReadFile{
			Filename: state.Statsfile,
			Success:  func(data []byte) Message { return StatsData{Data: data} },
		},
		PeriodicInterrupt{250 * time.Millisecond},
	)
}
