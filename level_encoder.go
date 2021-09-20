// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strings"

	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

/** Cached Level Encoder ***/

func init() {
	patt.DefaultEncoders.LevelEncoder = NewLevelEncoder("")
	patt.DefaultEncoders.BeginColorizer = NewBeginColorizer("")
	patt.DefaultEncoders.EndColorizer = NewEndColorizer("")
}

type cacheLevel struct {
	cache map[int][]byte

	color bool
	upper bool
	short bool
}

// NewLevelEncoder creates a new level encoder.
func NewLevelEncoder(typ string) patt.Encoder {
	e := &cacheLevel{}
	return e.NewEncoder(typ)
}

func (*cacheLevel) NewEncoder(typ string) patt.Encoder {
	e := &cacheLevel{}
	switch typ {
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
	return e
}

func (e *cacheLevel) Encode(out *bytes.Buffer, r *driver.Recorder) {
	n := r.Level
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

type colorLevel struct {
	isBegin bool
	encode  func(out *bytes.Buffer, n int)
}

// NewBeginColorizer creates a new color begin encoder.
func NewBeginColorizer(typ string) patt.Encoder {
	e := &colorLevel{isBegin: true}
	return e.NewEncoder(typ)
}

// NewEndColorizer creates a new color reset encoder.
func NewEndColorizer(typ string) patt.Encoder {
	e := &colorLevel{isBegin: false}
	return e.NewEncoder(typ)
}

func (e0 *colorLevel) NewEncoder(typ string) patt.Encoder {
	// Clear cache and remember mode
	e := &colorLevel{isBegin: e0.isBegin}

	isColor := false
	switch typ {
	case "color":
		isColor = true
	case "auto":
		isColor = color.IsTerminal()
	case "std":
		fallthrough
	default:
	}

	if !isColor {
		e.encode = e.Nop
		return e
	}

	if e.isBegin {
		e.encode = e.Begin
	} else {
		e.encode = e.End
	}
	return e
}

func (e *colorLevel) Encode(out *bytes.Buffer, r *driver.Recorder) {
	e.encode(out, r.Level)
}

func (e *colorLevel) Begin(out *bytes.Buffer, n int) {
	out.Write(Level(n).colorBytes(n))
}

func (e *colorLevel) End(out *bytes.Buffer, n int) {
	out.Write(color.Reset.Bytes())
}

func (e *colorLevel) Nop(out *bytes.Buffer, n int) {
}
