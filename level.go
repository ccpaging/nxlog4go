// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"fmt"
	"strings"

	color "github.com/ccpaging/nxlog4go/ansicolor"
	"github.com/ccpaging/nxlog4go/patt"
)

// logging levels used by the logger
const (
	FINEST int = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARN
	ERROR
	CRITICAL
)

type levelString struct {
	short string
	lower string
	color color.Color
}

var levelMap = map[int]*levelString{
	FINEST:   &levelString{"FNST", "finest", color.Gray},
	FINE:     &levelString{"FINE", "fine", color.Green},
	DEBUG:    &levelString{"DEBG", "debug", color.Magenta},
	TRACE:    &levelString{"TRAC", "trace", color.Cyan},
	INFO:     &levelString{"INFO", "info", color.White},
	WARN:     &levelString{"WARN", "warn", color.LightYellow},
	ERROR:    &levelString{"EROR", "error", color.Red},
	CRITICAL: &levelString{"CRIT", "critical", color.LightRed},
}

// Level is the integer logging levels
type Level int

// String return the string of integer Level
func (l Level) String() string {
	ls, ok := levelMap[int(l)]
	if ok {
		return ls.short
	}
	return fmt.Sprintf("Level(%d)", l)
}

func (l Level) string2int(s string) int {
	s = strings.ToLower(s)
	for i, ls := range levelMap {
		if s == strings.ToLower(ls.short) || s == ls.lower {
			return i
		}
	}
	if s == "warning" {
		return WARN
	}
	return int(l)
}

// IntE casts an interface to a level int.
func (l Level) IntE(v interface{}) (int, error) {
	if _, ok := v.(int); ok {
		return v.(int), nil
	}
	if _, ok := v.(Level); ok {
		return int(v.(Level)), nil
	}
	if _, ok := v.(string); ok {
		return l.string2int(v.(string)), nil
	}

	return INFO, fmt.Errorf("unknown level value %#v of type %T", v, v)
}

// Int casts an interface to a level int.
func (l Level) Int(v interface{}) int {
	n, _ := l.IntE(v)
	return n
}

// ColorBytes return the ANSI color bytes by level
func (l Level) colorBytes(n int) []byte {
	ls, ok := levelMap[int(n)]
	if ok {
		return ls.color.Bytes()
	}
	return color.Red.Bytes()
}

// Colorize return the ANSI color wrap bytes by level
func (l Level) colorize(s string) []byte {
	ls, ok := levelMap[int(l)]
	if ok {
		return ls.color.Wrap([]byte(s))
	}
	return color.Red.Wrap([]byte(s))
}

/** Cached Level Encoder ***/

type cacheLevel struct {
	cache map[int][]byte

	color bool
	upper bool
	short bool
}

func setLevelEncoder() {
	patt.Encoders.Level = &cacheLevel{}
}

func (e *cacheLevel) enco(out *bytes.Buffer, n int) {
	if e.cache == nil {
		e.cache = make(map[int][]byte, len(levelMap))
		for i, ls := range levelMap {
			s := ls.lower
			if e.short {
				s = ls.short
			} else if e.upper {
				s = strings.ToUpper(s)
			}

			if e.color {
				e.cache[i] = Level(i).colorize(s)
			} else {
				e.cache[i] = []byte(s)
			}
		}
	}

	if b, ok := e.cache[n]; ok {
		out.Write(b)
	} else {
		s := Level(n).String()
		if e.color {
			out.Write(Level(n).colorize(s))
		} else {
			out.Write([]byte(s))
		}
	}
}

// Open creates cached level encoding by name.
// Name includes(case sensitive): upper, upperColor, lower, lowerColor, std.
//
// Default: std.
func (*cacheLevel) Encoding(s string) patt.LevelEncoding {
	// make a new one
	e := new(cacheLevel)
	switch s {
	case "upper":
		e.upper = true
	case "upperColor":
		e.upper = true
		e.color = true
	case "lower":
	case "lowerColor":
		e.color = true
	case "std":
		fallthrough
	default:
		e.short = true
	}
	return e.enco
}

func (*cacheLevel) Begin(n int) []byte {
	return Level(n).colorBytes(n)
}

func (*cacheLevel) End(n int) []byte {
	return color.Reset.Bytes()
}
