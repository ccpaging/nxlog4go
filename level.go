// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"strings"

	"github.com/ccpaging/nxlog4go/ansicolor"
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

// Logging level strings
var levelStrings = map[int]string{
	FINEST:   "FNST",
	FINE:     "FINE",
	DEBUG:    "DEBG",
	TRACE:    "TRAC",
	INFO:     "INFO",
	WARN:     "WARN",
	ERROR:    "EROR",
	CRITICAL: "CRIT",
}

// Logging level color
var levelColors = map[int]color.Color{
	FINEST:   color.Gray,
	FINE:     color.Green,
	DEBUG:    color.Magenta,
	TRACE:    color.Cyan,
	INFO:     color.White,
	WARN:     color.LightYellow,
	ERROR:    color.Red,
	CRITICAL: color.LightRed,
}

// Logging level lowercase strings
var levelLowerStrings = map[int]string{
	FINEST:   "finest",
	FINE:     "fine",
	DEBUG:    "debug",
	TRACE:    "trace",
	INFO:     "info",
	WARN:     "warn",
	ERROR:    "error",
	CRITICAL: "critical",
}

// Level is the integer logging levels
type Level int

// String return the string of integer Level
func (l Level) String() string {
	s, ok := levelStrings[int(l)]
	if ok {
		return s
	}
	return l.unknown()
}

// Unknown return unknown level string
func (l Level) unknown() string {
	return fmt.Sprintf("Level(%d)", l)
}

// ColorBytes return the ANSI color bytes by level
func (l Level) ColorBytes() []byte {
	c, ok := levelColors[int(l)]
	if ok {
		return c.Bytes()
	}
	return color.Red.Bytes()
}

// ColorReset return the reset ANSI color bytes
func (l Level) ColorReset() []byte {
	return color.Reset.Bytes()
}

// Colorize return the ANSI color wrap bytes by level
func (l Level) Colorize(s string) []byte {
	c, ok := levelColors[int(l)]
	if ok {
		return c.Wrap([]byte(s))
	}
	return color.Red.Wrap([]byte(s))
}

// Int return the integer level of string
func (l Level) Int(s string) int {
	s = strings.ToLower(s)
	for i := 0; (i < len(levelStrings)) || (i < len(levelStrings)); i++ {
		if i < len(levelStrings) && s == levelStrings[i] {
			return i
		}
		if i < len(levelLowerStrings) && s == levelLowerStrings[i] {
			return i
		}
	}
	if s == "WARNING" {
		return WARN
	}
	return int(l)
}

// Max return maximum level int
func (l Level) Max() int {
	return CRITICAL
}
