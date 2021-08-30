// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"strings"

	"github.com/ccpaging/nxlog4go/color"
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

var levelMap map[int]*levelString

func init() {
	levelMap = make(map[int]*levelString)
	levelMap[FINEST] = &levelString{"FNST", "finest", color.Gray}
	levelMap[FINE] = &levelString{"FINE", "fine", color.Green}
	levelMap[DEBUG] = &levelString{"DEBG", "debug", color.Magenta}
	levelMap[TRACE] = &levelString{"TRAC", "trace", color.Cyan}
	levelMap[INFO] = &levelString{"INFO", "info", color.White}
	levelMap[WARN] = &levelString{"WARN", "warn", color.LightYellow}
	levelMap[ERROR] = &levelString{"EROR", "error", color.Red}
	levelMap[CRITICAL] = &levelString{"CRIT", "critical", color.LightRed}
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
