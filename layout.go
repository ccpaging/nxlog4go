// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strings"
	"sync"
	"time"
)

// Layout is is an interface for formatting log record
type Layout interface {
	// Set option about the Layout. The options should be set as default.
	// Chainable.
	Set(name string, v interface{}) Layout

	// Set option about the Layout. The options should be set as default.
	// Checkable
	SetOption(name string, v interface{}) error

	// This will be called to log a LogRecord message.
	Format(rec *LogRecord) []byte
}

var (
	// PatternDefault includes date, time, zone, level, source, lines, and message
	PatternDefault = "[%D %T %z] [%L] (%s:%N) %M\n"
	// PatternShort includes short time, short date, level and message
	PatternShort = "[%h:%m %d] [%L] %M\n"
	// PatternAbbrev includes level and message
	PatternAbbrev = "[%L] %M\n"
	// PatternJSON is json format include everyone of log record
	PatternJSON = "{\"Level\":%l,\"Created\":\"%YT%U%Z\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%N,\"Message\":\"%M\"}"
)

// PatternLayout formats log record with pattern
type PatternLayout struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	pattSlice [][]byte   // Split the pattern into pieces by % signs
	utc       bool
	longZone
	shortZone []byte
}

// NewPatternLayout creates a new layout which format log record by pattern.
// Using PatternDefault if pattern is empty string.
func NewPatternLayout(pattern string) Layout {
	if pattern == "" {
		pattern = PatternDefault
	}
	pl := &PatternLayout{}
	return pl.Set("pattern", pattern).Set("utc", false)
}

// Set option of layout. chainable
func (pl *PatternLayout) Set(k string, v interface{}) Layout {
	pl.SetOption(k, v)
	return pl
}

// SetOption set option with:
//	pattern	  - Layout format pattern
//	utc		  - Log recorder time zone
//	
// Known pattern codes are:
//	%U - Time (15:04:05.000000)
//	%T - Time (15:04:05)
//	%h - hour
//	%m - minute
//	%Z - Zone (-0700)
//	%z - Zone (MST)
//	%D - Date (2006/01/02)
//	%Y - Date (2006-01-02)
//	%d - Date (01/02/06)
//	%L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
//	%l - Level
//	%P - Prefix
//	%S - Source
//	%s - Short Source
//	%N - Line number
//	%M - Message
//	%t - Return (\t)
//	%r - Return (\r)
//	%n - Return (\n)
//	Ignores other unknown formats
func (pl *PatternLayout) SetOption(k string, v interface{}) (err error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	err = nil

	switch k {
	case "pattern", "format":
		if value, ok := v.(string); ok {
			pl.pattSlice = bytes.Split([]byte(value), []byte{'%'})
		} else if value, ok := v.([]byte); ok {
			pl.pattSlice = bytes.Split(value, []byte{'%'})
		} else {
			err = ErrBadValue
		}
	case "utc":
		utc := false
		if utc, err = ToBool(v); err == nil {
			t := time.Now()
			if utc {
				t = t.UTC()
			}
			pl.shortZone = []byte(t.Format("MST"))
			pl.longZone = []byte(t.Format("Z07:00"))
			pl.utc = utc
		}
	default:
		err = ErrBadOption
	}

	return
}

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

func formatHMS(out *bytes.Buffer, t *time.Time, sep byte) {
	hh, mm, ss := t.Clock()
	var b [16]byte
	b[0] = byte('0' + hh/10)
	b[1] = byte('0' + hh%10)
	b[2] = sep
	b[3] = byte('0' + mm/10)
	b[4] = byte('0' + mm%10)
	b[5] = sep
	b[6] = byte('0' + ss/10)
	b[7] = byte('0' + ss%10)
	out.Write(b[:8])
}

func formatDMY(out *bytes.Buffer, t *time.Time, sep byte) {
	y, m, d := t.Date()
	y %= 100
	var b [16]byte
	b[0] = byte('0' + d/10)
	b[1] = byte('0' + d%10)
	b[2] = sep
	b[3] = byte('0' + m/10)
	b[4] = byte('0' + m%10)
	b[5] = sep
	b[6] = byte('0' + y/10)
	b[7] = byte('0' + y%10)
	out.Write(b[:8])
}

func formatCYMD(out *bytes.Buffer, t *time.Time, sep byte) {
	y, m, d := t.Date()
	c := y / 100
	y %= 100
	var b [16]byte
	b[0] = byte('0' + c/10)
	b[1] = byte('0' + c%10)
	b[2] = byte('0' + y/10)
	b[3] = byte('0' + y%10)
	b[4] = sep
	b[5] = byte('0' + m/10)
	b[6] = byte('0' + m%10)
	b[7] = sep
	b[8] = byte('0' + d/10)
	b[9] = byte('0' + d%10)
	out.Write(b[:10])
}

func writeRecord(out *bytes.Buffer, piece0 byte, rec *LogRecord) {
	switch piece0 {
	case 'L':
		out.WriteString(levelStrings[rec.Level])
	case 'P':
		out.WriteString(rec.Prefix)
	case 'S':
		out.WriteString(rec.Source)
	case 's':
		out.WriteString(rec.Source[strings.LastIndex(rec.Source, "/")+1:])
	case 'l':
		var b []byte
		itoa(&b, int(rec.Level), -1)
		out.Write(b)
	case 'N':
		var b []byte
		itoa(&b, rec.Line, -1)
		out.Write(b)
	case 'M':
		out.WriteString(rec.Message)
	case 't':
		out.WriteByte('\t')
	case 'r':
		out.WriteByte('\r')
	case 'n', 'R':
		out.WriteByte('\n')
	}
}

func (pl *PatternLayout) writePiece(out *bytes.Buffer, piece []byte, rec *LogRecord) {
	// assert len(pieces) > 0
	t := rec.Created
	if pl.utc {
		t = t.UTC()
	}
	var b []byte
	switch piece[0] {
	case 'U':
		formatHMS(out, &t, ':')
		b = append(b, '.')
		itoa(&b, t.Nanosecond()/1e3, 6)
		out.Write(b)
	case 'T':
		formatHMS(out, &t, ':')
	case 'h':
		itoa(&b, t.Hour(), 2)
		out.Write(b)
	case 'm':
		itoa(&b, t.Minute(), 2)
		out.Write(b)
	case 'Z':
		out.Write(pl.longZone)
	case 'z':
		out.Write(pl.shortZone)
	case 'D':
		formatCYMD(out, &t, '/')
	case 'Y':
		formatCYMD(out, &t, '-')
	case 'd':
		formatDMY(out, &t, '/')
	default:
		writeRecord(out, piece[0], rec)
	}
	if len(piece) > 1 {
		out.Write(piece[1:])
	}
}

// Format log record.
// Return bytes.
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
	// Iterate over the pieces, replacing known formats
	// Split the string into pieces by % signs
	// pieces := bytes.Split([]byte(format), []byte{'%'})
	for i, piece := range pl.pattSlice {
		if i == 0 && len(piece) > 0 {
			out.Write(piece)
			continue
		}
		if len(piece) <= 0 {
			continue
		}
		pl.writePiece(out, piece, rec)
	}

	return out.Bytes()
}
