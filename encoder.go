// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (c) 2016 Uber Technologies, Inc.

package nxlog4go

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

/* Date Cache Encoder */
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

func (e *cacheTime) writeCacheDate(buf *bytes.Buffer, t *time.Time) bool {
	y, m, d := t.Date()
	if y == e.y && int(m) == e.mm && d == e.d && e.dateCache != nil {
		buf.Write(e.dateCache)
		return true
	}
	e.y, e.mm, e.d = y, int(m), d
	return false
}

func (e *cacheTime) writeDate(buf *bytes.Buffer, t *time.Time) {
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

func (e *cacheTime) writeCYMD(buf *bytes.Buffer, t *time.Time) {
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

// DateEncoder serializes a date to a []byte type.
type DateEncoder func(buf *bytes.Buffer, t *time.Time)

// NewDateEncoder creates cached date encoding by name.
// Name includes(case sensitive): dmy, mdy, cymdDash, cymdDot, cymdSlash.
// Default: cymdSlash.
func NewDateEncoder(s string) DateEncoder {
	e := new(cacheTime)
	switch s {
	case "dmy":
		e.dayFirst = true
		e.sep = '/'
		return e.writeDate
	case "mdy":
		e.sep = '/'
		return e.writeDate
	case "cymdDash":
		e.sep = '-'
	case "cymdDot":
		e.sep = '.'
	case "cymdSlash":
		fallthrough
	default:
		e.sep = '/'
	}
	return e.writeCYMD
}

/* Time Cache Encoder */
func (e *cacheTime) writeCacheTime(buf *bytes.Buffer, t *time.Time) bool {
	h, m, s := t.Clock()
	if s == e.s && m == e.m && h == e.h && e.timeCache != nil {
		buf.Write(e.timeCache)
		return true
	}
	e.h, e.m, e.s = h, m, s
	return false
}

func (e *cacheTime) writeHMS(buf *bytes.Buffer, t *time.Time) {
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
	e.writeCYMD(buf, t)

	buf.WriteByte('T')

	e.writeHMS(buf, t)

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

	e.writeZone(buf, t)
}

func (e *cacheTime) iso8601(buf *bytes.Buffer, t *time.Time) {
	// 2006-01-02T15:04:05.000Z0700
	e.writeCYMD(buf, t)

	buf.WriteByte('T')

	e.writeHMS(buf, t)

	buf.WriteByte('.')

	var b []byte
	itoa(&b, t.Nanosecond()/1e6, 3)
	buf.Write(b)

	e.writeZone(buf, t)
}

// TimeEncoder serializes a time to a []byte type.
type TimeEncoder func(buf *bytes.Buffer, t *time.Time)

// NewTimeEncoder creates cached time encoding by name.
// Name includes(case sensitive): hhmm, hms.us, iso88601, rfc3339nano, hms.
// Default: hms.
func NewTimeEncoder(s string) TimeEncoder {
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
	return e.writeHMS
}

/* Zone Encoder */

func (e *cacheTime) writeZone(buf *bytes.Buffer, t *time.Time) {
	loc := t.Location()
	if e.loc != loc || e.zoneCache == nil {
		e.zoneCache = []byte(t.Format(e.zfmt))
		e.loc = loc
	}
	buf.Write(e.zoneCache)
}

// ZoneEncoder serializes a time zone to a []byte type.
type ZoneEncoder func(buf *bytes.Buffer, t *time.Time)

// NewZoneEncoder creates cached time zone encoding by name.
// Name includes(case sensitive): rfc3339, iso88601, mst.
// Default: mst.
func NewZoneEncoder(s string) ZoneEncoder {
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
	return e.writeZone
}

/* Caller Encoder */

type callerEncoMode int

const (
	fullPath callerEncoMode = iota
	shortPath
	noPath
)

func (m callerEncoMode) write(buf *bytes.Buffer, s string) {
	if len(s) <= 0 {
		return
	}

	if m == fullPath {
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

	if m != noPath {
		// Find the penultimate separator.
		idx = strings.LastIndexByte(s[:idx], '/')
		if idx == -1 {
			buf.WriteString(s)
			return
		}
	}
	buf.WriteString(s[idx+1:])
}

// CallerEncoder function serializes a caller information to a []byte type.
type CallerEncoder func(buf *bytes.Buffer, s string)

// NewCallerEncoder creates caller encoder by name.
// Name includes(case sensitive): nopath, fullpath, shortpath.
// Default: shortpath.
func NewCallerEncoder(s string) CallerEncoder {
	var e CallerEncoder
	switch s {
	case "nopath":
		e = noPath.write
	case "fullpath":
		e = fullPath.write
	case "shortpath":
		fallthrough
	default:
		e = shortPath.write
	}
	return e
}

/* Fields Encoder */

type fieldsEncoMode struct {
	sep   string
	deli  string
	quote bool
}

func (e *fieldsEncoMode) writeKeyValue(out *bytes.Buffer, k string, v interface{}) {
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

func (e *fieldsEncoMode) write(out *bytes.Buffer, data map[string]interface{}, index []string) {
	if len(data) <= 0 {
		return
	}

	if len(index) > 1 {
		for _, k := range index {
			out.WriteString(e.sep)
			e.writeKeyValue(out, k, data[k])
		}
		return
	}

	for k, v := range data {
		out.WriteString(e.sep)
		e.writeKeyValue(out, k, v)
	}
}

func (e *fieldsEncoMode) writeJSON(out *bytes.Buffer, data map[string]interface{}, index []string) {
	if len(data) <= 0 {
		return
	}

	out.WriteString(",\"Data\":")
	encoder := json.NewEncoder(out)
	encoder.Encode(data)
}

// FieldsEncoder serializes data fields to a []byte type.
type FieldsEncoder func(out *bytes.Buffer, data map[string]interface{}, index []string)

// NewFieldsEncoder creates fields encoder by name.
// Name includes(case sensitive): csv, json, quote, std.
// Default: std.
func NewFieldsEncoder(s string) FieldsEncoder {
	e := new(fieldsEncoMode)
	switch s {
	case "json":
		return e.writeJSON
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
	return e.write
}
