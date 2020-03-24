// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (c) 2016 Uber Technologies, Inc.

package color

import (
	"os"
)

// Color represents a text color.
type Color uint8

// Foreground colors.
const (
	Black Color = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Gray
	LightRed
	LightGreen
	LightYellow
	LightBlue
	LightMagenta
	LightCyan
	LightWhite
	Reset
)

var colorBytes = map[Color][]byte{
	Black:   []byte("\033[30m"),
	Red:     []byte("\033[31m"),
	Green:   []byte("\033[32m"),
	Yellow:  []byte("\033[33m"),
	Blue:    []byte("\033[34m"),
	Magenta: []byte("\033[35m"),
	Cyan:    []byte("\033[36m"),
	White:   []byte("\033[37m"),

	Gray:         []byte("\033[30;1m"),
	LightRed:     []byte("\033[31;1m"),
	LightGreen:   []byte("\033[32;1m"),
	LightYellow:  []byte("\033[33;1m"),
	LightBlue:    []byte("\033[34;1m"),
	LightMagenta: []byte("\033[35;1m"),
	LightCyan:    []byte("\033[36;1m"),
	LightWhite:   []byte("\033[37;1m"),
	Reset:        []byte("\033[0m"),
}

func IsTerminal() bool {
	return (os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb") || os.Getenv("ConEmuANSI") == "ON"
}

// Wrap the coloring to the given bytes and return.
func (c Color) Wrap(s []byte) []byte {
	if b, ok := colorBytes[c]; ok {
		b = append(b, s...)
		b = append(b, colorBytes[Reset]...)
		return b
	}
	return s
}

// Bytes return ANSI color bytes
func (c Color) Bytes() []byte {
	if color, ok := colorBytes[c]; ok {
		return color
	}
	return []byte{}
}
