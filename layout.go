// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strings"
	"time"
	"sync"
)

// This is an interface for formatting log record
type Layout interface {
	// Set option about the Layout. The options should be set as default.
	// Chainable.
	Set(name string, v interface{}) Layout

	// Set option about the Layout. The options should be set as default.
	// Checkable
	SetOption(name string, v interface{}) error

	// Get(name string) string

	// This will be called to log a LogRecord message.
	Format(rec *LogRecord) []byte
}

var (
	PATTERN_DEFAULT = "[%D %T %z] [%L] (%s:%N) %M\n"
	PATTERN_SHORT   = "[%h:%m %d] [%L] %M\n"
	PATTERN_ABBREV  = "[%L] %M\n"
	PATTERN_JSON	= "{\"Level\":%l,\"Created\":\"%YT%U%Z\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%N,\"Message\":\"%M\"}"
)

// This layout formats log record by pattern
type PatternLayout struct {
	mu sync.Mutex // ensures atomic writes; protects the following fields
	pattSlice [][]byte // Split the pattern into pieces by % signs
	utc bool
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
func (pl *PatternLayout) Set(name string, v interface{}) Layout {
	pl.SetOption(name, v)
	return pl
}

/* 
Set option. checkable. Better be set before the first log message is written.
Known pattern codes:
	%U - Time (15:04:05.000000)
	%T - Time (15:04:05)
	%h - hour
	%m - minute
	%Z - Zone (-0700)
	%z - Zone (MST)
	%D - Date (2006/01/02)
	%Y - Date (2006-01-02)
	%d - Date (01/02/06)
	%L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
	%l - Level
	%P - Prefix
	%S - Source
	%s - Short Source
	%N - Line number
	%M - Message
	%t - Return (\t)
	%r - Return (\r)
	%n - Return (\n)
	Ignores unknown formats
Recommended: "[%D %T] [%L] (%S) %M"
*/
func (pl *PatternLayout) SetOption(name string, v interface{}) (err error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	err = nil

	switch name {
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

/*
// Get option
func (pl PatternLayout) Get(name string) string {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if name == "pattern" {
		return string(bytes.Join(pl.pattSlice, []byte{'%'}))
	}
	return ""
}
*/

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

func formatHHMMSS(buf *[]byte, hh, mm, ss int) {
	var b [16]byte
	b[0] = byte('0' + hh / 10); b[1] = byte('0' + hh % 10)
	b[2] = byte(':')
	b[3] = byte('0' + mm / 10); b[4] = byte('0' + mm % 10)
	b[5] = byte(':')
	b[6] = byte('0' + ss / 10); b[7] = byte('0' + ss % 10)
	*buf = append(*buf, b[:8]...)
}

func formatCCYYMMDD(buf *[]byte, cc, yy, mm, dd int, sep byte) {
	var b [16]byte
	b[0] = byte('0' + cc / 10); b[1] = byte('0' + cc % 10)
	b[2] = byte('0' + yy / 10); b[3] = byte('0' + yy % 10)
	b[4] = sep
	b[5] = byte('0' + mm / 10); b[6] = byte('0' + mm % 10)
	b[7] = sep
	b[8] = byte('0' + dd / 10); b[9] = byte('0' + dd % 10)
	*buf = append(*buf, b[:10]...)
}

func formatDDMMYY(buf *[]byte, yy, mm, dd int) {
	var b [16]byte
	b[0] = byte('0' + dd / 10); b[1] = byte('0' + dd % 10)
	b[2] = byte('/')
	b[3] = byte('0' + mm / 10); b[4] = byte('0' + mm % 10)
	b[5] = byte('/')
	b[6] = byte('0' + yy / 10); b[7] = byte('0' + yy % 10)
	*buf = append(*buf, b[:8]...)
}

func writeRecord(out *bytes.Buffer, piece0 byte, rec *LogRecord) {
	switch piece0 {
	case 'L': out.WriteString(levelStrings[rec.Level])
	case 'P': out.WriteString(rec.Prefix)
	case 'S': out.WriteString(rec.Source)
	case 's': out.WriteString(rec.Source[strings.LastIndex(rec.Source, "/")+1:])
	case 'M': out.WriteString(rec.Message)
	}
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
	if pl.utc { t = t.UTC() }

	year, month, day := t.Date(); hour, minute, second := t.Clock()
	// Split the string into pieces by % signs
	//pieces := bytes.Split([]byte(format), []byte{'%'})
	var b []byte
	// Iterate over the pieces, replacing known formats
	for i, piece := range pl.pattSlice {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'U':
				formatHHMMSS(&b, hour, minute, second)
				b = append(b, '.')
				itoa(&b, t.Nanosecond()/1e3, 6)
			case 'T': formatHHMMSS(&b, hour, minute, second)
			case 'h': itoa(&b, hour, 2)
			case 'm': itoa(&b, minute, 2)
			case 'Z': out.Write(pl.longZone)
			case 'z': out.Write(pl.shortZone)
			case 'D': formatCCYYMMDD(&b, year / 100, year % 100, int(month), int(day), '/')
			case 'Y': formatCCYYMMDD(&b, year / 100, year % 100, int(month), int(day), '-')
			case 'd': formatDDMMYY(&b, year % 100, int(month), int(day))
			case 't': out.WriteByte('\t')
			case 'r': out.WriteByte('\r')
			case 'n', 'R': out.WriteByte('\n')
			case 'l': itoa(&b, int(rec.Level), -1)
			case 'N': itoa(&b, rec.Line, -1)
			default:
				writeRecord(out, piece[0], rec)
			}
			if len(b) > 0 {
				out.Write(b)
				b = nil
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
