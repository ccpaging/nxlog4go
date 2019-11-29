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

/* Level Cache Encoder */
type cacheLevel struct {
	cache map[int][]byte
}

func (ca *cacheLevel) write(out *bytes.Buffer, level int, format string, isColor bool) {
	if ca.cache == nil {
		ls := levelLowerStrings
		isUpper := false
		switch format {
		case "upper":
			isUpper = true
		case "upperColor":
			isUpper = true
		case "lower":
		case "lowerColor":
		case "std":
			fallthrough
		default:
			ls = levelStrings
		}

		ca.cache = make(map[int][]byte, len(ls))
		for i, s := range ls {
			if isUpper {
				s = strings.ToUpper(s)
			}
			if isColor {
				ca.cache[i] = levelColors[i].Wrap([]byte(s))
			} else {
				ca.cache[i] = []byte(s)
			}
		}
	}

	if b, ok := ca.cache[level]; ok {
		out.Write(b)
	} else {
		s := Level(level).Unknown()
		if isColor {
			out.Write(Red.Wrap([]byte(s)))
		} else {
			out.Write([]byte(s))
		}
	}
}

func (ca *cacheLevel) std(out *bytes.Buffer, level int) {
	ca.write(out, level, "std", false)
}

func (ca *cacheLevel) lower(out *bytes.Buffer, level int) {
	ca.write(out, level, "lower", false)
}

func (ca *cacheLevel) lowerColor(out *bytes.Buffer, level int) {
	ca.write(out, level, "lowerColor", true)
}

func (ca *cacheLevel) upper(out *bytes.Buffer, level int) {
	ca.write(out, level, "upper", false)
}

func (ca *cacheLevel) upperColor(out *bytes.Buffer, level int) {
	ca.write(out, level, "upperColor", true)
}

// LevelEncoder serializes a Level to a []byte type.
type LevelEncoder func(buf *bytes.Buffer, n int)

// NewLevelEncoder creates cached level encoding by name.
// Name includes(case sensitive): upper, upperColor, lower, lowerColor, std.
// Default: std.
func NewLevelEncoder(s string) LevelEncoder {
	ca := new(cacheLevel)
	switch s {
	case "upper":
		return ca.upper
	case "upperColor":
		return ca.upperColor
	case "lower":
		return ca.lower
	case "lowerColor":
		return ca.lowerColor
	case "std":
		fallthrough
	default:
	}
	return ca.std
}

/* Date Cache Encoder */
type cacheTime struct {
	y, mm, d  int
	dateCache []byte

	h, m, s   int // the number of seconds elapsed since January 1, 1970 UTC
	timeCache []byte

	loc       *time.Location
	zoneCache []byte
}

func (ca *cacheTime) writeDateCache(buf *bytes.Buffer, t *time.Time) bool {
	y, m, d := t.Date()
	if y == ca.y && int(m) == ca.mm && d == ca.d && ca.dateCache != nil {
		buf.Write(ca.dateCache)
		return true
	}
	ca.y, ca.mm, ca.d = y, int(m), d
	return false
}

func (ca *cacheTime) dmy(buf *bytes.Buffer, t *time.Time) {
	if ca.writeDateCache(buf, t) {
		return
	}

	y, m, d := t.Date()
	y %= 100

	var b [16]byte
	b[0] = byte('0' + d/10)
	b[1] = byte('0' + d%10)
	b[2] = '/'
	b[3] = byte('0' + m/10)
	b[4] = byte('0' + m%10)
	b[5] = '/'
	b[6] = byte('0' + y/10)
	b[7] = byte('0' + y%10)

	ca.dateCache = b[:8]

	buf.Write(ca.dateCache)
}

func (ca *cacheTime) mdy(buf *bytes.Buffer, t *time.Time) {
	if ca.writeDateCache(buf, t) {
		return
	}

	y, m, d := t.Date()
	y %= 100

	var b [16]byte
	b[0] = byte('0' + m/10)
	b[1] = byte('0' + m%10)
	b[2] = '/'
	b[3] = byte('0' + d/10)
	b[4] = byte('0' + d%10)
	b[5] = '/'
	b[6] = byte('0' + y/10)
	b[7] = byte('0' + y%10)

	ca.dateCache = b[:8]

	buf.Write(ca.dateCache)
}

func (ca *cacheTime) cymd(buf *bytes.Buffer, t *time.Time, sep byte) {
	if ca.writeDateCache(buf, t) {
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
	b[4] = sep
	b[5] = byte('0' + m/10)
	b[6] = byte('0' + m%10)
	b[7] = sep
	b[8] = byte('0' + d/10)
	b[9] = byte('0' + d%10)

	ca.dateCache = b[:10]

	buf.Write(ca.dateCache)
}

func (ca *cacheTime) cymdDash(buf *bytes.Buffer, t *time.Time) {
	ca.cymd(buf, t, '-')
}

func (ca *cacheTime) cymdSlash(buf *bytes.Buffer, t *time.Time) {
	ca.cymd(buf, t, '/')
}

func (ca *cacheTime) cymdDot(buf *bytes.Buffer, t *time.Time) {
	ca.cymd(buf, t, '.')
}

// DateEncoder serializes a date to a []byte type.
type DateEncoder func(buf *bytes.Buffer, t *time.Time)

// NewDateEncoder creates cached date encoding by name.
// Name includes(case sensitive): dmy, mdy, cymdDash, cymdDot, cymdSlash.
// Default: cymdSlash.
func NewDateEncoder(s string) DateEncoder {
	ca := new(cacheTime)
	switch s {
	case "dmy":
		return ca.dmy
	case "mdy":
		return ca.mdy
	case "cymdDash":
		return ca.cymdDash
	case "cymdDot":
		return ca.cymdDot
	case "cymdSlash":
		fallthrough
	default:
	}
	return ca.cymdSlash
}

/* Time Cache Encoder */
func (ca *cacheTime) writeTimeCache(buf *bytes.Buffer, t *time.Time) bool {
	h, m, s := t.Clock()
	if h == ca.h && m == ca.m && s == ca.s && ca.timeCache != nil {
		buf.Write(ca.timeCache)
		return true
	}
	ca.h, ca.m, ca.s = h, m, s
	return false
}

func (ca *cacheTime) hhmm(buf *bytes.Buffer, t *time.Time) {
	if ca.writeTimeCache(buf, t) {
		return
	}

	hh, mm, _ := t.Clock()

	var b [16]byte
	b[0] = byte('0' + hh/10)
	b[1] = byte('0' + hh%10)
	b[2] = ':'
	b[3] = byte('0' + mm/10)
	b[4] = byte('0' + mm%10)

	ca.timeCache = b[:5]
	buf.Write(ca.timeCache)
}

func (ca *cacheTime) hms(buf *bytes.Buffer, t *time.Time) {
	if ca.writeTimeCache(buf, t) {
		return
	}

	hh, mm, ss := t.Clock()
	var b [16]byte
	b[0] = byte('0' + hh/10)
	b[1] = byte('0' + hh%10)
	b[2] = ':'
	b[3] = byte('0' + mm/10)
	b[4] = byte('0' + mm%10)
	b[5] = ':'
	b[6] = byte('0' + ss/10)
	b[7] = byte('0' + ss%10)

	ca.timeCache = b[:8]
	buf.Write(ca.timeCache)
}

func (ca *cacheTime) hmsMicrosecond(buf *bytes.Buffer, t *time.Time) {
	ca.hms(buf, t)

	var b []byte
	b = append(b, '.')
	itoa(&b, t.Nanosecond()/1e3, 6)
	buf.Write(b)
}

func (ca *cacheTime) rfc3339Nano(buf *bytes.Buffer, t *time.Time) {
	// 2006-01-02T15:04:05.000000Z07:00
	ca.cymdDash(buf, t)

	buf.WriteByte('T')

	ca.hms(buf, t)

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

	ca.zoneRFC3339(buf, t)
}

func (ca *cacheTime) iso8601(buf *bytes.Buffer, t *time.Time) {
	// 2006-01-02T15:04:05.000Z0700
	ca.cymdDash(buf, t)

	buf.WriteByte('T')

	ca.hms(buf, t)

	buf.WriteByte('.')

	var b []byte
	itoa(&b, t.Nanosecond()/1e6, 3)
	buf.Write(b)

	ca.zoneISO8601(buf, t)
}

// TimeEncoder serializes a time to a []byte type.
type TimeEncoder func(buf *bytes.Buffer, t *time.Time)

// NewTimeEncoder creates cached time encoding by name.
// Name includes(case sensitive): hhmm, hms.us, iso88601, rfc3339nano, hms.
// Default: hms.
func NewTimeEncoder(s string) TimeEncoder {
	ca := new(cacheTime)
	switch s {
	case "hhmm":
		return ca.hhmm
	case "hms.us":
		return ca.hmsMicrosecond
	case "iso8601":
		return ca.iso8601
	case "rfc3339nano":
		return ca.rfc3339Nano
	case "hms":
		fallthrough
	default:
	}
	return ca.hms
}

/* Zone Encoder */

func (ca *cacheTime) writeZoneCache(buf *bytes.Buffer, t *time.Time) bool {
	loc := t.Location()
	if loc == ca.loc && ca.zoneCache != nil {
		buf.Write(ca.zoneCache)
		return true
	}
	ca.loc = loc
	return false
}

func (ca *cacheTime) mst(buf *bytes.Buffer, t *time.Time) {
	if ca.writeZoneCache(buf, t) {
		return
	}

	ca.zoneCache = []byte(t.Format("MST"))
	buf.Write(ca.zoneCache)
}

func (ca *cacheTime) zoneRFC3339(buf *bytes.Buffer, t *time.Time) {
	if ca.writeZoneCache(buf, t) {
		return
	}

	ca.zoneCache = []byte(t.Format("Z07:00"))
	buf.Write(ca.zoneCache)
}

func (ca *cacheTime) zoneISO8601(buf *bytes.Buffer, t *time.Time) {
	if ca.writeZoneCache(buf, t) {
		return
	}

	ca.zoneCache = []byte(t.Format("Z0700"))
	buf.Write(ca.zoneCache)
}

// ZoneEncoder serializes a time zone to a []byte type.
type ZoneEncoder func(buf *bytes.Buffer, t *time.Time)

// NewZoneEncoder creates cached time zone encoding by name.
// Name includes(case sensitive): rfc3339, iso88601, mst.
// Default: mst.
func NewZoneEncoder(s string) ZoneEncoder {
	ca := new(cacheTime)
	switch s {
	case "rfc3339":
		return ca.zoneRFC3339
	case "iso8601":
		return ca.zoneISO8601
	case "mst":
		fallthrough
	default:
	}
	return ca.mst
}

/* Caller Encoder */

func fullpathCallerEncoder(buf *bytes.Buffer, s string) {
	if len(s) <= 0 {
		return
	}
	buf.WriteString(s)
}

func shortCallerEncoder(buf *bytes.Buffer, s string) {
	if len(s) <= 0 {
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
	// Find the penultimate separator.
	idx = strings.LastIndexByte(s[:idx], '/')
	if idx == -1 {
		buf.WriteString(s)
		return
	}
	buf.WriteString(s[idx+1:])
}

func nopathCallerEncoder(buf *bytes.Buffer, s string) {
	if len(s) <= 0 {
		return
	}
	idx := strings.LastIndexByte(s, '/')
	if idx == -1 {
		buf.WriteString(s)
		return
	}
	buf.WriteString(s[idx+1:])
}

// CallerEncoder function serializes a caller information to a []byte type.
type CallerEncoder func(buf *bytes.Buffer, s string)

// SetAs set caller encoder by name.
// Name includes(case sensitive): nopath, fullpath, shortpath.
// Default: shortpath.
func (e *CallerEncoder) SetAs(s string) {
	switch s {
	case "nopath":
		*e = nopathCallerEncoder
	case "fullpath":
		*e = fullpathCallerEncoder
	case "shortpath":
		fallthrough
	default:
		*e = shortCallerEncoder
	}
}

/* Fields Encoder */
func writeKeyValue(buf *bytes.Buffer, k string, v interface{}) {
	buf.WriteString(k + "=")
	var b []byte
	if _, ok := v.(string); ok {
		b = []byte(v.(string))
	} else {
		b = []byte(fmt.Sprint(v))
	}
	if len(b) <= 16 && bytes.IndexAny(b, " =") < 0 {
		// no quote
	} else {
		b = append([]byte{'"'}, b...)
		b = append(b, byte('"'))
	}
	buf.Write(b)
}

func writeFileds(buf *bytes.Buffer, data map[string]interface{}, index []string, sep []byte) {
	if len(data) <= 0 {
		return
	}

	if len(index) > 1 {
		for _, k := range index {
			buf.Write(sep)
			writeKeyValue(buf, k, data[k])
		}
		return
	}

	for k, v := range data {
		buf.Write(sep)
		writeKeyValue(buf, k, v)
	}
}

func csvFieldsEncoder(buf *bytes.Buffer, data map[string]interface{}, index []string) {
	writeFileds(buf, data, index, []byte{'|'})
}

func keyvalFieldsEncoder(buf *bytes.Buffer, data map[string]interface{}, index []string) {
	writeFileds(buf, data, index, []byte{' '})
}

func jsonFieldsEncoder(buf *bytes.Buffer, data map[string]interface{}, index []string) {
	if len(data) <= 0 {
		return
	}

	buf.WriteString(",\"Data\":")
	encoder := json.NewEncoder(buf)
	encoder.Encode(data)
}

// FieldsEncoder serializes data fields to a []byte type.
type FieldsEncoder func(buf *bytes.Buffer, data map[string]interface{}, index []string)

// SetAs set fields encoder by name.
// Name includes(case sensitive): csv, json, keyval.
// Default: keyval.
func (e *FieldsEncoder) SetAs(s string) {
	switch s {
	case "csv":
		*e = csvFieldsEncoder
	case "json":
		*e = jsonFieldsEncoder
	case "keyval":
		fallthrough
	default:
		*e = keyvalFieldsEncoder
	}
}
