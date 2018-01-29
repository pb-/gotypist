# Gotypist [![Build Status](https://travis-ci.org/pb-/gotypist.svg?branch=master)](https://travis-ci.org/pb-/gotypist)

A simple touch-typing tutor that follows [Steve Yegge's methodology](http://steve-yegge.blogspot.com/2008/09/programmings-dirtiest-little-secret.html) of going in fast, slow, and medium cycles.

![Screenshot of a Gotypist session, normal mode](screenshot.png)

The TL;DR of this methodology is that you make three passes for each phrase.

 * **Fast** - type as fast as you can without any attention to proper technique or correctness
 * **Slow** - use proper technique at all costs, go as slow as needed to achieve that
 * **Normal** - type at *target speed*, try to make no mistakes

Gotypist will score each pass accordingly: errors do not matter in the first pass, speed does not matter in the second pass, and both matter in the final pass. Check out `score.go` for details (which are subject to change).

This project was mainly motivated by trying out [termbox-go](https://github.com/nsf/termbox-go), but it is definitely ready for productive learning.

## Installation

```
go get github.com/pb-/gotypist
```

## Usage

    gotypist [-f FILE] [-s] [-n PROB] [-c] [WORD]...

    WORD...     Explicitly specify a phrase
    -f FILE     Use FILE instead of /usr/share/dict/words as data source
    -n PROB     Sprinkle in random numbers with probability 0 <= PROB <= 1
    -c          Tread -f FILE as code and go sequenntially through the lines
    -d          Run in demo mode to take a screenshot

## Key bindings

    ESC   quit
    C-F   skip forward to the next phrase
    C-R   toggle repeat phrase mode
    C-I   toggle finger usage hints

## Code organization

The code loosely follows an [Elm-like architecture](https://guide.elm-lang.org/architecture/). In a nutshell that means all interesting and Gotypist-specific code resides within pure functions. This is quite experimental and some corners were cut since Go is not primarily a functional language, but it still enjoys a lot of the benefits of this architectural style!
