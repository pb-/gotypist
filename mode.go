package main

import "github.com/nsf/termbox-go"

type Mode int

const (
	ModeFast Mode = iota
	ModeSlow
	ModeNormal
)

var modeInfo = []struct {
	Name string
	Desc string
	Attr termbox.Attribute
}{{
	Name: "fast",
	Desc: "type as fast as you can, ignore mistakes",
	Attr: termbox.ColorGreen | termbox.AttrBold,
}, {
	Name: "slow",
	Desc: "go slow, do not make any mistake",
	Attr: termbox.ColorMagenta | termbox.AttrBold,
}, {
	Name: "normal",
	Desc: "type at normal speed, avoid mistakes",
	Attr: termbox.ColorYellow | termbox.AttrBold,
}}

func (m Mode) Num() int {
	return int(m)
}

func (m Mode) Name() string {
	return modeInfo[m].Name
}

func (m Mode) Desc() string {
	return modeInfo[m].Desc
}

func (m Mode) Attr() termbox.Attribute {
	return modeInfo[m].Attr
}
