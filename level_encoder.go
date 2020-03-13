// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strings"

	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/patt"
)

/** Cached Level Encoder ***/

func init() {
	stde := patt.GetEncoders()
	stde.Level = &cacheLevel{}
	patt.SetEncoders(stde)
}

type cacheLevel struct {
	cache map[int][]byte

	color bool
	upper bool
	short bool
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
