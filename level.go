// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"strings"
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
var levelColors = map[int]Color{
	FINEST:   Gray,
	FINE:     Green,
	DEBUG:    Magenta,
	TRACE:    Cyan,
	INFO:     White,
	WARN:     LightYellow,
	ERROR:    Red,
	CRITICAL: LightRed,
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
	return l.Unknown()
}

// Unknown return unknown level string
func (l Level) Unknown() string {
	return fmt.Sprintf("Level(%d)", l)
}

// Color return the ANSI color bytes of level
func (l Level) Color() []byte {
	c, ok := levelColors[int(l)]
	if ok {
		return c.Bytes()
	}
	return Red.Bytes()
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
