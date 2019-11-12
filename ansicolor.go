// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (c) 2016 Uber Technologies, Inc.

package nxlog4go

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
	ResetColor
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
	ResetColor:   []byte("\033[0m"),
}

// Add the coloring to the given bytes and return.
func (c Color) Wrap(s []byte) []byte {
	if b, ok := colorBytes[c]; ok {
		b = append(b, s...)
		b = append(b, colorBytes[ResetColor]...)
		return b
	}
	return s
}

func (c Color) Bytes() []byte {
	if color, ok := colorBytes[c]; ok {
		return color
	}
	return []byte{}
}

func setColor(e *Entry) bool {
	if out := e.logger.out; out != nil {
		switch e.Level {
		case CRITICAL:
			out.Write(LightRed.Bytes())
		case ERROR:
			out.Write(Red.Bytes())
		case WARN:
			out.Write(LightYellow.Bytes())
		case INFO:
		case TRACE:
			out.Write(Magenta.Bytes())
		case DEBUG:
			out.Write(Green.Bytes())
		case FINE:
			out.Write(Cyan.Bytes())
		case FINEST:
			out.Write(Blue.Bytes())
		default:
		}
	}
	return true
}

func resetColor(e *Entry, n int, err error) {
	if out := e.logger.out; out != nil {
		out.Write(ResetColor.Bytes())
	}
}
