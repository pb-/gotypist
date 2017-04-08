# Gotypist

A simple touch-typing tutor that follows [Steve Yegge's methodology](http://steve-yegge.blogspot.com/2008/09/programmings-dirtiest-little-secret.html) of going in fast, slow, and medium cycles.

![Screenshot of a Gotypist session, normal mode](screenshot.png)

This project is mainly motivated by trying out [termbox-go](https://github.com/nsf/termbox-go), but it is definitely ready for productive learning.

## Usage

    gotypist [-w FILE] [-s] [WORD]...

    -w FILE     Use this file as word list instead of /usr/share/dict/words
    -s          Run in demo mode to take a screenshot
    WORD...     Explicitly specify a phrase

## Key bindings

    ESC   quit
    C-F   skip forward to the next phrase
    C-R   toggle repeat phrase mode
