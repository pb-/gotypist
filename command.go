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

type AppendFile struct {
	Filename string
	Data     []byte
	Success  func() Message
	Error    func(error) Message
}

type Exit struct {
	Status         int
	GoodbyeMessage string
}

func PassError(err error) Message {
	return err
}

func RunCommand(cmd Command) []Message {
	switch c := cmd.(type) {
	case ReadFile:
		return readFile(c.Filename, c.Success, c.Error)
	case AppendFile:
		return appendFile(c.Filename, c.Data, c.Success, c.Error)
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

func appendFile(filename string, data []byte, success func() Message, error func(error) Message) []Message {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		if error != nil {
			return []Message{error(err)}
		}
		return noMessages
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		if error != nil {
			return []Message{error(err)}
		}
		return noMessages
	}

	if success == nil {
		return noMessages
	}

	return []Message{success()}
}

func readFile(filename string, success func([]byte) Message, errorFunc func(error) Message) []Message {
	var (
		content []byte
		err     error
	)

	if filename == "-" {
		content, err = ioutil.ReadAll(os.Stdin)
	} else {
		content, err = ioutil.ReadFile(filename)
	}

	if err != nil {
		if errorFunc != nil {
			return []Message{errorFunc(err)}
		}
		return noMessages
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
