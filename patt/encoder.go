// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (c) 2016 Uber Technologies, Inc.

package patt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ccpaging/nxlog4go/driver"
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

/** Encoder **/

// Encoder defines log recorder field encoder interface for external packages extending.
type Encoder interface {
	// Open opens a new Encoder according type.
	Open(typ string) Encoder
	// Encode serializes log recorder field to the bytes buffer.
	Encode(out *bytes.Buffer, r *driver.Recorder)
}

type nopEncoder struct{}

// NewNopEncoder creates no-op encoder.
func NewNopEncoder() Encoder                                 { return &nopEncoder{} }
func (e *nopEncoder) Open(string) Encoder                    { return e }
func (e *nopEncoder) Encode(*bytes.Buffer, *driver.Recorder) {}

/** cached Time Encoder **/

const (
	modeTime int = iota
	modeDate
	modeZone
)

type cacheTime struct {
	mode   int
	encode func(buf *bytes.Buffer, t *time.Time)

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

// NewDateEncoder creates a new date encoder.
func NewDateEncoder(typ string) Encoder {
	e := &cacheTime{mode: modeDate}
	return e.Open(typ)
}

// NewTimeEncoder creates a new time encoder.
func NewTimeEncoder(typ string) Encoder {
	e := &cacheTime{mode: modeTime}
	return e.Open(typ)
}

// NewZoneEncoder creates a new time zone encoder.
func NewZoneEncoder(typ string) Encoder {
	e := &cacheTime{mode: modeZone}
	return e.Open(typ)
}

func (e *cacheTime) Open(typ string) Encoder {
	// Clear cache and keep mode
	ne := &cacheTime{mode: e.mode}
	switch ne.mode {
	case modeDate:
		ne.setDate(typ)
	case modeZone:
		ne.setZone(typ)
	case modeTime:
		fallthrough
	default:
		ne.setTime(typ)
	}
	return ne
}

func (e *cacheTime) Encode(out *bytes.Buffer, r *driver.Recorder) {
	e.encode(out, &r.Created)
}

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

func (e *cacheTime) setDate(typ string) {
	switch typ {
	case "dmy":
		e.dayFirst = true
		e.sep = '/'
		e.encode = e.encoDate
		return
	case "mdy":
		e.sep = '/'
		e.encode = e.encoDate
		return
	case "cymdDash":
		e.sep = '-'
	case "cymdDot":
		e.sep = '.'
	case "cymdSlash":
		fallthrough
	default:
		e.sep = '/'
	}
	e.encode = e.encoCYMD
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

func (e *cacheTime) setTime(s string) {
	switch s {
	case "iso8601":
		e.century = true
		e.sep = '-'
		e.zfmt = "Z0700"
		e.encode = e.iso8601
		return
	case "rfc3339nano":
		e.century = true
		e.sep = '-'
		e.zfmt = "Z07:00"
		e.encode = e.rfc3339Nano
		return
	case "hhmm":
		e.nos = true
	case "hms.us":
		e.us = true
	case "hms":
		fallthrough
	default:
	}
	e.encode = e.encoHMS
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

func (e *cacheTime) setZone(s string) {
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
	e.encode = e.encoZone
}

/* Caller Encoder */

type callerEncoder struct {
	mode int
}

const (
	fullPath int = iota
	shortPath
	noPath
)

// NewCallerEncoder creates a new caller path encoder.
func NewCallerEncoder(typ string) Encoder {
	e := &callerEncoder{}
	return e.Open(typ)
}

func (*callerEncoder) Open(typ string) Encoder {
	e := &callerEncoder{}
	switch typ {
	case "nopath":
		e.mode = noPath
	case "fullpath":
		e.mode = fullPath
	case "shortpath":
		fallthrough
	default:
		e.mode = shortPath
	}
	return e
}

func (e *callerEncoder) Encode(buf *bytes.Buffer, r *driver.Recorder) {
	s := r.Source

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

/* Fields Encoder */
type fieldsEncoder struct {
	sep   string
	deli  string
	quote bool

	encode func(out *bytes.Buffer, fields map[string]interface{}, index []string)
}

// NewFieldsEncoder creates a new fields encoder.
func NewFieldsEncoder(typ string) Encoder {
	e := &fieldsEncoder{}
	return e.Open(typ)
}

func (*fieldsEncoder) Open(typ string) Encoder {
	e := &fieldsEncoder{}
	switch typ {
	case "json":
		e.encode = e.encoJSON
		return e
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
	e.encode = e.encoStd
	return e
}

func (e *fieldsEncoder) Encode(out *bytes.Buffer, r *driver.Recorder) {
	e.encode(out, r.Fields, r.Index)
}

func (e *fieldsEncoder) encoKeyValue(out *bytes.Buffer, k string, v interface{}) {
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

func (e *fieldsEncoder) encoStd(out *bytes.Buffer, fields map[string]interface{}, index []string) {
	if len(fields) <= 0 {
		return
	}

	if len(index) > 1 {
		for _, k := range index {
			out.WriteString(e.sep)
			e.encoKeyValue(out, k, fields[k])
		}
		return
	}

	for k, v := range fields {
		out.WriteString(e.sep)
		e.encoKeyValue(out, k, v)
	}
}

func (e *fieldsEncoder) encoJSON(out *bytes.Buffer, fields map[string]interface{}, index []string) {
	if len(fields) <= 0 {
		return
	}

	out.WriteString(",\"Fields\":")
	encoder := json.NewEncoder(out)
	encoder.Encode(fields)
}

/* Values Encoder */
type valuesEncoder struct {
	encode func(out *bytes.Buffer, values []interface{})
}

// NewValuesEncoder creates a new data fields encoder.
func NewValuesEncoder(typ string) Encoder {
	e := &valuesEncoder{}
	return e.Open(typ)
}

func (e *valuesEncoder) Encode(out *bytes.Buffer, r *driver.Recorder) {
	e.encode(out, r.Values)
}

func (*valuesEncoder) Open(typ string) Encoder {
	e := &valuesEncoder{}
	e.encode = e.encoStd
	return e
}

func (e *valuesEncoder) encoStd(out *bytes.Buffer, values []interface{}) {
	if len(values) <= 0 {
		return
	}
	out.WriteString(fmt.Sprint(values...))
}
