// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (c) 2016 Uber Technologies, Inc.

package patt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

/** Level Encoder **/

// LevelEncoding serializes a Level to a bytes buffer.
type LevelEncoding func(out *bytes.Buffer, n int)

// LevelEncoder defines level encoder interface for external packages extending.
type LevelEncoder interface {
	Encoding(string) LevelEncoding
	Begin(int) []byte
	End(int) []byte
}

type nopLevelEnc struct{}

// NewNopLevelEncoder creates no-op level encoder.
func NewNopLevelEncoder() LevelEncoder                 { return &nopLevelEnc{} }
func (e *nopLevelEnc) Encoding(s string) LevelEncoding { return e.enco }
func (*nopLevelEnc) enco(out *bytes.Buffer, n int)     {}
func (*nopLevelEnc) Begin(n int) []byte                { return nil }
func (*nopLevelEnc) End(n int) []byte                  { return nil }

/** Time Encoder **/

// TimeEncoding serializes date, time, or zone to bytes buffer.
type TimeEncoding func(buf *bytes.Buffer, t *time.Time)

// TimeEncoder defines date encoder interface for external packages extending.
type TimeEncoder interface {
	DateEncoding(string) TimeEncoding
	TimeEncoding(string) TimeEncoding
	ZoneEncoding(string) TimeEncoding
}

type cacheTime struct {
	y, mm, d  int
	dateCache []byte
	dayFirst  bool // true, dmy; false, mdy
	sep       byte // '-', '/', '.'
	century   bool

	h, m, s   int // the number of seconds elapsed since January 1, 1970 UTC
	timeCache []byte
	nos       bool
	us        bool

	loc       *time.Location
	zoneCache []byte
	zfmt      string
}

// NewTimeEncoder creates the default time encoder.
func NewTimeEncoder() TimeEncoder { return &cacheTime{} }

/** Date Encoding **/

func (e *cacheTime) writeCacheDate(buf *bytes.Buffer, t *time.Time) bool {
	y, m, d := t.Date()
	if y == e.y && int(m) == e.mm && d == e.d && e.dateCache != nil {
		buf.Write(e.dateCache)
		return true
	}
	e.y, e.mm, e.d = y, int(m), d
	return false
}

func (e *cacheTime) encoDate(buf *bytes.Buffer, t *time.Time) {
	if e.writeCacheDate(buf, t) {
		return
	}

	y, m, d := t.Date()
	y %= 100

	var b [16]byte
	i := 0
	if e.dayFirst {
		b[i] = byte('0' + d/10)
		b[i+1] = byte('0' + d%10)
		b[i+2] = e.sep
		i += 3
	}
	b[i] = byte('0' + m/10)
	b[i+1] = byte('0' + m%10)
	b[i+2] = e.sep
	i += 3
	if !e.dayFirst {
		b[i] = byte('0' + d/10)
		b[i+1] = byte('0' + d%10)
		b[i+2] = e.sep
		i += 3
	}
	b[i] = byte('0' + y/10)
	b[i+1] = byte('0' + y%10)
	i += 2

	e.dateCache = b[:i]

	buf.Write(e.dateCache)
}

func (e *cacheTime) encoCYMD(buf *bytes.Buffer, t *time.Time) {
	if e.writeCacheDate(buf, t) {
		return
	}

	y, m, d := t.Date()
	c := y / 100
	y %= 100

	var b [16]byte
	b[0] = byte('0' + c/10)
	b[1] = byte('0' + c%10)
	b[2] = byte('0' + y/10)
	b[3] = byte('0' + y%10)
	b[4] = e.sep
	b[5] = byte('0' + m/10)
	b[6] = byte('0' + m%10)
	b[7] = e.sep
	b[8] = byte('0' + d/10)
	b[9] = byte('0' + d%10)

	e.dateCache = b[:10]

	buf.Write(e.dateCache)
}

// DateEncoding creates cached date encoding by name.
// Name includes(case sensitive): dmy, mdy, cymdDash, cymdDot, cymdSlash.
//
// Default: cymdSlash.
func (*cacheTime) DateEncoding(s string) TimeEncoding {
	e := new(cacheTime)
	switch s {
	case "dmy":
		e.dayFirst = true
		e.sep = '/'
		return e.encoDate
	case "mdy":
		e.sep = '/'
		return e.encoDate
	case "cymdDash":
		e.sep = '-'
	case "cymdDot":
		e.sep = '.'
	case "cymdSlash":
		fallthrough
	default:
		e.sep = '/'
	}
	return e.encoCYMD
}

/** Time Encoding **/

func (e *cacheTime) writeCacheTime(buf *bytes.Buffer, t *time.Time) bool {
	h, m, s := t.Clock()
	if s == e.s && m == e.m && h == e.h && e.timeCache != nil {
		buf.Write(e.timeCache)
		return true
	}
	e.h, e.m, e.s = h, m, s
	return false
}

func (e *cacheTime) encoHMS(buf *bytes.Buffer, t *time.Time) {
	if e.writeCacheTime(buf, t) {
		return
	}

	hh, mm, ss := t.Clock()

	var b [16]byte
	b[0] = byte('0' + hh/10)
	b[1] = byte('0' + hh%10)
	b[2] = ':'
	b[3] = byte('0' + mm/10)
	b[4] = byte('0' + mm%10)

	if e.nos {
		e.timeCache = b[:5]
		buf.Write(e.timeCache)
		return
	}

	b[5] = ':'
	b[6] = byte('0' + ss/10)
	b[7] = byte('0' + ss%10)

	e.timeCache = b[:8]
	buf.Write(e.timeCache)

	if e.us {
		var us []byte
		us = append(us, '.')
		itoa(&us, t.Nanosecond()/1e3, 6)
		buf.Write(us)
	}
}

func (e *cacheTime) rfc3339Nano(buf *bytes.Buffer, t *time.Time) {
	// 2006-01-02T15:04:05.000000Z07:00
	e.encoCYMD(buf, t)

	buf.WriteByte('T')

	e.encoHMS(buf, t)

	var b []byte
	itoa(&b, t.Nanosecond(), 9)

	// trim '0'
	n := len(b)
	for n > 0 && b[n-1] == '0' {
		n--
	}

	if n > 0 {
		buf.WriteByte('.')
		buf.Write(b[:n])
	}

	e.encoZone(buf, t)
}

func (e *cacheTime) iso8601(buf *bytes.Buffer, t *time.Time) {
	// 2006-01-02T15:04:05.000Z0700
	e.encoCYMD(buf, t)

	buf.WriteByte('T')

	e.encoHMS(buf, t)

	buf.WriteByte('.')

	var b []byte
	itoa(&b, t.Nanosecond()/1e6, 3)
	buf.Write(b)

	e.encoZone(buf, t)
}

// TimeEncoding creates cached time encoding by name.
// Name includes(case sensitive): hhmm, hms.us, iso88601, rfc3339nano, hms.
//
// Default: hms.
func (*cacheTime) TimeEncoding(s string) TimeEncoding {
	e := new(cacheTime)
	switch s {
	case "iso8601":
		e.century = true
		e.sep = '-'
		e.zfmt = "Z0700"
		return e.iso8601
	case "rfc3339nano":
		e.century = true
		e.sep = '-'
		e.zfmt = "Z07:00"
		return e.rfc3339Nano
	case "hhmm":
		e.nos = true
	case "hms.us":
		e.us = true
	case "hms":
		fallthrough
	default:
	}
	return e.encoHMS
}

/* Zone Encoder */

func (e *cacheTime) encoZone(buf *bytes.Buffer, t *time.Time) {
	loc := t.Location()
	if e.loc != loc || e.zoneCache == nil {
		e.zoneCache = []byte(t.Format(e.zfmt))
		e.loc = loc
	}
	buf.Write(e.zoneCache)
}

// ZoneEncoding creates cached time zone encoding by name.
// Name includes(case sensitive): rfc3339, iso88601, mst.
//
// Default: mst.
func (*cacheTime) ZoneEncoding(s string) TimeEncoding {
	e := new(cacheTime)
	switch s {
	case "iso8601":
		e.zfmt = "Z0700"
	case "rfc3339":
		e.zfmt = "Z07:00"
	case "mst":
		fallthrough
	default:
		e.zfmt = "MST"
	}
	return e.encoZone
}

/* Caller Encoder */

// CallerEncoding function serializes a caller information to a bytes buffer.
type CallerEncoding func(buf *bytes.Buffer, s string)

// CallerEncoder defines caller encoder interface for external packages extending.
type CallerEncoder interface {
	Encoding(string) CallerEncoding
}

type callerEnc struct {
	mode int
}

const (
	fullPath int = iota
	shortPath
	noPath
)

// NewCallerEncoder creates the default caller information encoder.
func NewCallerEncoder() CallerEncoder { return &callerEnc{} }

func (e *callerEnc) enco(buf *bytes.Buffer, s string) {
	if len(s) <= 0 {
		return
	}

	if e.mode == fullPath {
		buf.WriteString(s)
		return
	}

	// nb. To make sure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.
	//
	// Find the last separator.
	//
	idx := strings.LastIndexByte(s, '/')
	if idx == -1 {
		buf.WriteString(s)
		return
	}

	if e.mode != noPath {
		// Find the penultimate separator.
		idx = strings.LastIndexByte(s[:idx], '/')
		if idx == -1 {
			buf.WriteString(s)
			return
		}
	}
	buf.WriteString(s[idx+1:])
}

// Encoding creates caller encoding by name.
// Name includes(case sensitive): nopath, fullpath, shortpath.
//
// Default: shortpath.
func (*callerEnc) Encoding(s string) CallerEncoding {
	e := new(callerEnc)
	switch s {
	case "nopath":
		e.mode = noPath
	case "fullpath":
		e.mode = fullPath
	case "shortpath":
		fallthrough
	default:
		e.mode = shortPath
	}
	return e.enco
}

/* Fields Encoder */

// FieldsEncoding serializes data fields to a bytes buffer.
type FieldsEncoding func(out *bytes.Buffer, data map[string]interface{}, index []string)

// FieldsEncoder defines data fields encoder interface for external packages extending.
type FieldsEncoder interface {
	Encoding(string) FieldsEncoding
}

type fieldsEnc struct {
	sep   string
	deli  string
	quote bool
}

// NewFieldsEncoder creates the default data fields encoder.
func NewFieldsEncoder() FieldsEncoder { return &fieldsEnc{} }

func (e *fieldsEnc) encoKeyValue(out *bytes.Buffer, k string, v interface{}) {
	out.WriteString(k + e.deli)
	var s string
	if e.quote {
		s = fmt.Sprintf("%q", v)
	} else {
		if _, ok := v.(string); ok {
			s = v.(string)
		} else {
			s = fmt.Sprint(v)
		}
	}
	out.WriteString(s)
}

func (e *fieldsEnc) enco(out *bytes.Buffer, data map[string]interface{}, index []string) {
	if len(data) <= 0 {
		return
	}

	if len(index) > 1 {
		for _, k := range index {
			out.WriteString(e.sep)
			e.encoKeyValue(out, k, data[k])
		}
		return
	}

	for k, v := range data {
		out.WriteString(e.sep)
		e.encoKeyValue(out, k, v)
	}
}

func (e *fieldsEnc) encoJSON(out *bytes.Buffer, data map[string]interface{}, index []string) {
	if len(data) <= 0 {
		return
	}

	out.WriteString(",\"Data\":")
	encoder := json.NewEncoder(out)
	encoder.Encode(data)
}

// Encoding creates a data fields encoder by name.
// Name includes(case sensitive): csv, json, quote, std.
//
// Default: std.
func (*fieldsEnc) Encoding(s string) FieldsEncoding {
	e := new(fieldsEnc)
	switch s {
	case "json":
		return e.encoJSON
	case "csv":
		e.sep = "|"
		e.deli = "="
	case "quote":
		e.sep = " "
		e.deli = "="
		e.quote = true
	case "std":
		fallthrough
	default:
		e.sep = " "
		e.deli = "="
	}
	return e.enco
}

/** Global Encoders ***/

type encoders struct {
	Level  LevelEncoder
	Time   TimeEncoder
	Caller CallerEncoder
	Fields FieldsEncoder
}

// Encoders is global encoders for external packages extending.
var Encoders encoders = encoders{
	Level:  NewNopLevelEncoder(),
	Time:   NewTimeEncoder(),
	Caller: NewCallerEncoder(),
	Fields: NewFieldsEncoder(),
}
