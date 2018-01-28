// contains all commands (side effects)
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

type Message interface{}

type Command interface{}

var Noop = []Command{}

type Interrupt struct {
	Delay time.Duration
}

type PeriodicInterrupt struct {
	Period time.Duration
}

type ReadFile struct {
	Filename string
	Success  func([]byte) Message
	Error    func(error) Message
}

type Exit struct {
	Status         int
	GoodbyeMessage string
}

func RunCommand(cmd Command) []Message {
	switch c := cmd.(type) {
	case ReadFile:
		return readFile(c.Filename, c.Success, c.Error)
	case Interrupt:
		return interrupt(c.Delay)
	case PeriodicInterrupt:
		return periodicInterrupt(c.Period)
	case Exit:
		return exit(c.Status, c.GoodbyeMessage)
	}

	exit(1, fmt.Sprintf("Cannot handle command of type %T", cmd))
	return noMessages
}

var noMessages = []Message{}

func readFile(filename string, success func([]byte) Message, error func(error) Message) []Message {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return []Message{error(err)}
	}
	return []Message{success(content)}
}

func periodicInterrupt(d time.Duration) []Message {
	go func() {
		for range time.Tick(d) {
			termbox.Interrupt()
		}
	}()

	return noMessages
}

func interrupt(d time.Duration) []Message {
	time.AfterFunc(d, termbox.Interrupt)
	return noMessages
}

func exit(status int, message string) []Message {
	termbox.Close()

	if message != "" {
		fmt.Println(message)
	}

	os.Exit(status)

	return noMessages
}
