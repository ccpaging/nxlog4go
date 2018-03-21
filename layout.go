// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strings"
	"time"
	"sync"
)

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// This is an interface for formatting log record
type Layout interface {
	// Set option about the Layout. The options should be set as default.
	// Chainable.
	Set(name string, v interface{}) Layout

	// Set option about the Layout. The options should be set as default.
	// Checkable
	SetOption(name string, v interface{}) error

	Get(name string) string

	// This will be called to log a LogRecord message.
	Format(rec *LogRecord) []byte
}

var (
	PATTERN_DEFAULT = "[%D %T %z] [%L] (%s:%n) %M\n"
	PATTERN_SHORT   = "[%t %d] [%L] %M\n"
	PATTERN_ABBREV  = "[%L] %M\n"
	PATTERN_JSON	= "{\"Level\":%l,\"Created\":\"%YT%N%Z\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%n,\"Message\":\"%M\"}"
)

// This layout formats log record by pattern
type PatternLayout struct {
	mu sync.Mutex // ensures atomic writes; protects the following fields
	pattSlice [][]byte // Split the pattern into pieces by % signs
	isUTC bool
	longZone, shortZone []byte
}

// NewPatternLayout creates a new layout which format log record by pattern
func NewPatternLayout(pattern string) Layout {
	if pattern == "" {
		pattern = PATTERN_DEFAULT
	}
	pl := &PatternLayout{}
	return pl.Set("pattern", pattern).Set("utc", false)
}

// Set option. chainable
// Known pattern codes:
// %N - Time (15:04:05.000000)
// %T - Time (15:04:05)
// %t - Time (15:04)
// %Z - Zone (-0700)
// %z - Zone (MST)
// %D - Date (2006/01/02)
// %Y - Date (2006-01-02)
// %d - Date (01/02/06)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %l - Level
// %P - Prefix
// %S - Source
// %s - Short Source
// %n - Line number
// %M - Message
// %R - Return (\n)
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
func (pl *PatternLayout) Set(name string, v interface{}) Layout {
	pl.SetOption(name, v)
	return pl
}

// Set option. checkable. Must be set before the first log message is written.
func (pl *PatternLayout) SetOption(name string, v interface{}) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	switch name {
	case "pattern", "format":
		if value, ok := v.(string); ok {
			pl.pattSlice = bytes.Split([]byte(value), []byte{'%'})
		} else if value, ok := v.([]byte); ok {
			pl.pattSlice = bytes.Split(value, []byte{'%'})
		} else {
			return ErrBadValue
		}
	case "utc":
		if value, ok := v.(string); ok {
			if value == "true" {
				pl.isUTC = true
			} else {
				pl.isUTC = false
			}
		} else if value, ok := v.(bool); ok {
			pl.isUTC = value
		} else {
			return ErrBadValue
		}
		t := time.Now()
		if pl.isUTC {
			t = t.UTC()
		}
		pl.shortZone = []byte(t.Format("MST"))
		pl.longZone = []byte(t.Format("Z07:00"))
	default:
		return ErrBadOption
	}

	return nil
}

// Get option
func (pl PatternLayout) Get(name string) string {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if name == "pattern" {
		return string(bytes.Join(pl.pattSlice, []byte{'%'}))
	}
	return ""
}

// Format log record
func (pl *PatternLayout) Format(rec *LogRecord) []byte {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if rec == nil {
		return []byte("<nil>")
	}
	if len(pl.pattSlice) == 0 {
		return nil
	}

	out := bytes.NewBuffer(make([]byte, 0, 64))

	t := rec.Created
	if pl.isUTC {
		t = t.UTC()
	}

	// Split the string into pieces by % signs
	//pieces := bytes.Split([]byte(format), []byte{'%'})
	var b []byte
	// Iterate over the pieces, replacing known formats
	for i, piece := range pl.pattSlice {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'N':
				b = nil
				hour, minute, second := t.Clock()
				itoa(&b, hour, 2); b = append(b, ':');
 				itoa(&b, minute, 2); b = append(b, ':');
 				itoa(&b, second, 2); b = append(b, '.')
				itoa(&b, t.Nanosecond()/1e3, 6)
				out.Write(b)
			case 'T':
				hour, minute, second := t.Clock()
				b = nil
				itoa(&b, hour, 2); b = append(b, ':');
 				itoa(&b, minute, 2); b = append(b, ':');
 				itoa(&b, second, 2)
				out.Write(b)
			case 't':
				hour, minute, _ := t.Clock()
				b = nil
				itoa(&b, hour, 2); b = append(b, ':'); itoa(&b, minute, 2)
				out.Write(b)
			case 'Z':
				out.Write(pl.longZone)
			case 'z':
				out.Write(pl.shortZone)
			case 'D':
				year, month, day := t.Date()
				b = nil
				itoa(&b, year, 4); b = append(b, '/')
				itoa(&b, int(month), 2); b = append(b, '/')
				itoa(&b, day, 2)
				out.Write(b)
			case 'Y':
				year, month, day := t.Date()
				b = nil
				itoa(&b, year, 4); b = append(b, '-')
				itoa(&b, int(month), 2); b = append(b, '-')
				itoa(&b, day, 2)
				out.Write(b)
			case 'd':
				year, month, day := t.Date()
				b = nil
				itoa(&b, day, 2); b = append(b, '/')
				itoa(&b, int(month), 2); b = append(b, '/')
				itoa(&b, year%100, 2)
				out.Write(b)
			case 'L':
				out.WriteString(levelStrings[rec.Level])
			case 'l':
				b = nil; itoa(&b, int(rec.Level), -1)
				out.Write(b)
			case 'P':
                out.WriteString(rec.Prefix)
			case 'S':
				out.WriteString(rec.Source)
			case 's':
				out.WriteString(rec.Source[strings.LastIndexByte(rec.Source, '/')+1:])
			case 'n':
				b = nil; itoa(&b, rec.Line, -1)
				out.Write(b)
			case 'M':
				out.WriteString(rec.Message)
			case 'R':
				out.WriteByte('\n')
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}

	return out.Bytes()
}
