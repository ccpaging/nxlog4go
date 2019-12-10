// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (c) 2016 Uber Technologies, Inc.

package nxlog4go

import (
	"bytes"
	"strings"
)

/* Level Cache Encoder */
type cacheLevel struct {
	cache map[int][]byte

	color bool
	upper bool
	short bool
}

func (e *cacheLevel) write(out *bytes.Buffer, level int) {
	if e.cache == nil {
		ls := levelLowerStrings
		if e.short {
			ls = levelStrings
		}

		e.cache = make(map[int][]byte, len(ls))
		for i, s := range ls {
			if e.upper {
				s = strings.ToUpper(s)
			}
			if e.color {
				e.cache[i] = Level(i).Colorize(s)
			} else {
				e.cache[i] = []byte(s)
			}
		}
	}

	if b, ok := e.cache[level]; ok {
		out.Write(b)
	} else {
		s := Level(level).String()
		if e.color {
			out.Write(Level(level).Colorize(s))
		} else {
			out.Write([]byte(s))
		}
	}
}

// LevelEncoder serializes a Level to a []byte type.
type LevelEncoder func(buf *bytes.Buffer, level int)

// NewLevelEncoder creates cached level encoding by name.
// Name includes(case sensitive): upper, upperColor, lower, lowerColor, std.
// Default: std.
func NewLevelEncoder(s string) LevelEncoder {
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
	return e.write
}
