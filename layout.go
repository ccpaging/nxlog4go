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

type timeCacheType struct {
	sync.RWMutex   // ensures atomic writes; protects the following fields
	LastUpdateSeconds   int64
	longTime, shortTime []byte
	longDate, shortDate []byte
	isUTC bool
	longZone, shortZone []byte
}
var timeCache = &timeCacheType{}

func (tc *timeCacheType) SetUTC(isUTC bool) {
	tc.Lock()
	defer tc.Unlock()
	t := time.Now()
	if isUTC {
		t = t.UTC()
	}
	tc.isUTC = isUTC
	tc.shortZone = []byte(t.Format("MST"))
	tc.longZone = []byte(t.Format("-0700"))
}

func (tc *timeCacheType) Update(t *time.Time) *timeCacheType {
	tc.Lock()
	defer tc.Unlock()
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	// fmt.Sprintf("%02d:%02d", hour, minute)
	tc.shortTime = nil
	buf := &tc.shortTime
	itoa(buf, hour, 2); *buf = append(*buf, ':')
	itoa(buf, minute, 2)
	// fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
	tc.longTime = nil
	buf = &tc.longTime
	*buf = append(*buf, tc.shortTime...); *buf = append(*buf, ':')
	itoa(buf, second, 2)
	// fmt.Sprintf("%02d/%02d/%02d", day, month, year%100)
	tc.shortDate = nil
	buf = &tc.shortDate
	itoa(buf, day, 2); *buf = append(*buf, '/')
	itoa(buf, int(month), 2); *buf = append(*buf, '/')
	itoa(buf, year%100, 2)
	// fmt.Sprintf("%04d/%02d/%02d", year, month, day),
	tc.longDate = nil
	buf = &tc.longDate
	itoa(buf, year, 4); *buf = append(*buf, '/')
	itoa(buf, int(month), 2); *buf = append(*buf, '/')
	itoa(buf, day, 2)
	return tc
}

func SetZoneUTC(isUTC bool) {
	timeCache.SetUTC(isUTC)
}

func init() {
	SetZoneUTC(false)
}

// This is an interface for anything that should be able to write logs
type Layout interface {
	// Set option about the Layout. The options should be set as default.
	// Must be set before the first log message is written if changed.
	// You should test more if have to change options while running.
	Set(name string, v interface{}) Layout

	Get(name string) string

	// This will be called to log a LogRecord message.
	Format(rec *LogRecord) []byte
}

var (
	PATTERN_DEFAULT = "[%D %T %z] [%L] (%s) %M"
	PATTERN_SHORT   = "[%t %d] [%L] %M"
	PATTERN_ABBREV  = "[%L] %M"
)

type PatternLayout struct {
	mu  sync.Mutex // ensures atomic writes; protects the following fields
	pattSlice [][]byte // Split the pattern into pieces by % signs
}

func NewPatternLayout(pattern string) Layout {
	if pattern == "" {
		pattern = PATTERN_DEFAULT
	}
	pl := &PatternLayout{}
	return pl.Set("pattern", pattern)
}

// Known format codes:
// %N - Time (15:04:05.000000)
// %T - Time (15:04:05)
// %t - Time (15:04)
// %Z - Zone (-0700)
// %z - Zone (MST)
// %D - Date (2006/01/02)
// %d - Date (01/02/06)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %P - Prefix
// %S - Source
// %s - Short Source
// %x - Extra Short Source: just file without .go suffix
// %M - Message
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
func (pl *PatternLayout) Set(name string, v interface{}) Layout {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if name == "pattern" {
		if pattern, ok := v.(string); ok {
			pl.pattSlice = bytes.Split([]byte(pattern), []byte{'%'})
		} else if pattern, ok := v.([]byte); ok {
			pl.pattSlice = bytes.Split(pattern, []byte{'%'})
		}
	}
	return pl
}

func (pl PatternLayout) Get(name string) string {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if name == "pattern" {
		return string(bytes.Join(pl.pattSlice, []byte{'%'}))
	}
	return ""
}

func (pl *PatternLayout) Format(rec *LogRecord) []byte {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if rec == nil {
		return []byte("<nil>")
	}
	if len(pl.pattSlice) == 0 {
		return nil
	}

	t := rec.Created
	if timeCache.isUTC {
		t = t.UTC()
	}

	out := bytes.NewBuffer(make([]byte, 0, 64))

	secs := t.UnixNano() / 1e9
	if timeCache.LastUpdateSeconds != secs {
		timeCache.LastUpdateSeconds = secs
		timeCache.Update(&t)
	}

	// Split the string into pieces by % signs
	//pieces := bytes.Split([]byte(format), []byte{'%'})
	var b []byte
	// Iterate over the pieces, replacing known formats
	timeCache.RLock()
	defer timeCache.RUnlock()
	for i, piece := range pl.pattSlice {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'N':
				out.Write(timeCache.longTime)
				b = nil; b = append(b, '.')
				itoa(&b, t.Nanosecond()/1e3, 6)
				out.Write(b)
			case 'T':
				out.Write(timeCache.longTime)
			case 't':
				out.Write(timeCache.shortTime)
			case 'Z':
				out.Write(timeCache.longZone)
			case 'z':
				out.Write(timeCache.shortZone)
			case 'D':
				out.Write(timeCache.longDate)
			case 'd':
				out.Write(timeCache.shortDate)
			case 'L':
				out.WriteString(levelStrings[rec.Level])
			case 'P':
                out.WriteString(rec.Prefix)
			case 'S':
				out.WriteString(rec.Source)
				if rec.Line > 0 {
					b = nil; b = append(b, ':')
					itoa(&b, rec.Line, -1)
					out.Write(b)
				}
			case 's':
				out.WriteString(rec.Source[strings.LastIndexByte(rec.Source, '/')+1:])
				if rec.Line > 0 {
					b = nil; b = append(b, ':')
					itoa(&b, rec.Line, -1)
					out.Write(b)
				}
			case 'M':
				out.WriteString(rec.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.Bytes()
}

type JsonLayout struct {
}

func NewJsonLayout() *JsonLayout {
	return &JsonLayout{}
}

func (jl *JsonLayout) Format(rec *LogRecord) []byte {
	out := bytes.NewBuffer(make([]byte, 0, 64))
	b := make([]byte, 0, 16)

	out.WriteString("{\"Level\":")
	out.WriteByte(byte('0' + int(rec.Level) % 10))
	out.WriteString(",")

	// 2018-01-26T02:05:55.3620165+08:00
	out.WriteString("\"Created\":")
	timeJson, _ := rec.Created.MarshalJSON()
	out.Write(timeJson)
	out.WriteString(",")

	out.WriteString("\"Prefix\":")
	out.WriteString("\"")
	out.WriteString(rec.Prefix)
	out.WriteString("\"")
	out.WriteString(",")

	out.WriteString("\"Source\":")
	out.WriteString("\"")
	out.WriteString(rec.Source)
	out.WriteString("\"")
	out.WriteString(",")

	out.WriteString("\"Line\":")
	b = nil; itoa(&b, rec.Line, -1)
	out.Write(b)
	out.WriteString(",")

	out.WriteString("\"Message\":")
	out.WriteString("\"")
	out.WriteString(rec.Message)
	out.WriteString("\"")
	out.WriteString("}")
	return out.Bytes()
}
