// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package patt

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/ccpaging/nxlog4go/cast"
	"github.com/ccpaging/nxlog4go/driver"
)

var (
	// DefaultLineEnd defines the default line ending when writing logs.
	// Alternate line endings specified in Encoder can override this
	// behavior.
	DefaultLineEnd = []byte("\n")
	// FormatDefault includes date, time, zone, level, source, lines, and message
	FormatDefault = "[%D %T %Z] [%L] (%S:%N) %M%F"
	// FormatConsole includes time, source, level and message
	FormatConsole = "%T %L (%S:%N) %M%F"
	// FormatShort includes short time, short date, level and message
	FormatShort = "[%T %D] [%L] %M%F"
	// FormatAbbrev includes level and message
	FormatAbbrev = "[%L] %M%F"
)

// PatternLayout formats log Recorder.
type PatternLayout struct {
	verbs   [][]byte // Split the format string into pieces by % signs
	lineEnd []byte

	utc   bool
	color bool

	LevelEncoding

	CallerEncoding
	DateEncoding TimeEncoding
	TimeEncoding
	ZoneEncoding TimeEncoding
	FieldsEncoding

	// DEPRECATED. Compatible with log4go
	_encodeDate TimeEncoding
	_encodeTime TimeEncoding // DEPRECATED
}

func formatToVerbs(format string) [][]byte {
	if format == "" {
		format = FormatDefault
	}
	if unq, err := strconv.Unquote(format); err == nil {
		format = unq
	}
	return bytes.Split([]byte(format), []byte{'%'})
}

// NewLayout creates a new layout encoding log Recorder.
func NewLayout(format string, args ...interface{}) *PatternLayout {
	lo := &PatternLayout{
		verbs:   formatToVerbs(format),
		lineEnd: DefaultLineEnd,

		utc: false,

		LevelEncoding:  Encoders.Level.Encoding(""),
		CallerEncoding: Encoders.Caller.Encoding("shortpath"),
		DateEncoding:   Encoders.Time.DateEncoding("cymdSlash"),
		TimeEncoding:   Encoders.Time.TimeEncoding("hms"),
		ZoneEncoding:   Encoders.Time.ZoneEncoding("mst"),
		FieldsEncoding: Encoders.Fields.Encoding("std"),

		// DEPRECATED. Compatible with log4go
		_encodeDate: Encoders.Time.DateEncoding("mdy"),
		_encodeTime: Encoders.Time.TimeEncoding("hhmm"),
	}
	lo.SetOptions(args...)
	return lo
}

// SetOptions sets name-value pair options.
//
// Return Layout interface.
func (lo *PatternLayout) SetOptions(args ...interface{}) driver.Layout {
	ops, idx, _ := driver.ArgsToMap(args)
	for _, k := range idx {
		lo.Set(k, ops[k])
	}
	return lo
}

func (lo *PatternLayout) setEncoder(k string, v interface{}) (err error) {
	var s string
	switch k {
	case "levelEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.LevelEncoding = Encoders.Level.Encoding(s)
		}
	case "callerEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.CallerEncoding = Encoders.Caller.Encoding(s)
		}
	case "dateEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.DateEncoding = Encoders.Time.DateEncoding(s)
		}
	case "timeEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.TimeEncoding = Encoders.Time.TimeEncoding(s)
		}
	case "zoneEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.ZoneEncoding = Encoders.Time.ZoneEncoding(s)
		}
	case "fieldsEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.FieldsEncoding = Encoders.Fields.Encoding(s)
		}
	default:
		return fmt.Errorf("unknown option name %s, value %#v of type %T", k, v, v)
	}

	return
}

// Set sets name-value option with:
//  format  - Layout format string. Auto-detecting quote string.
//  lineEnd - line end string. Auto-detecting quote string.
//  utc     - Log record time zone: local or utc.
//
// Known encoder types are (The option's name and value are case-sensitive):
//  levelEncoder  - "upper", "upperColor", "lower", "lowerColor", "std" is default.
//  callerEncoder - "nopath", "fullpath", "shortpath" is default.
//  dateEncoder   - "dmy", "mdy", "cymdDash", "cymdDot", "cymdSlash" is default.
//  timeEncoder   - "hhmm",  "hms.us", "iso8601", "rfc3339nano", "hms" is default.
//  zoneEncoder   - "rfc3339", "iso8601", "mst" is default.
//  fieldsEncoder - "quote", "csv", "json", "quote", "std" is default.
//
// Known format codes are:
//  %D - Date (2006/01/02)
//  %T - Time (15:04:05)
//  %Z - Zone (-0700)
//  %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
//  %l - Integer level
//  %P - Prefix
//  %S - Source
//  %N - Line number
//  %M - Message
//  %F - Data fields in "key=value" format
//
// DEPRECATED:
//  %d - Date (01/02/06). Replacing with setting "dateEncoder" as "mdy".
//  %t - Time (15:04). Replacing with setting "timeEncoder" as "hhmm".
//
// Ignores other unknown format codes
func (lo *PatternLayout) Set(k string, v interface{}) (err error) {
	var (
		s  string
		ok bool
	)
	switch k {
	case "format", "pattern":
		if s, err = cast.ToString(v); err == nil && len(s) > 0 {
			lo.verbs = formatToVerbs(s)
		}
	case "lineEnd":
		if s, err = cast.ToString(v); err == nil {
			var u string
			if u, err = strconv.Unquote(s); err == nil {
				s = u
			}
			lo.lineEnd = []byte(s)
		}
	case "color":
		if ok, err = cast.ToBool(v); err == nil {
			lo.color = ok
		}
	case "utc":
		if ok, err = cast.ToBool(v); err == nil {
			lo.utc = ok
		}
	default:
		return lo.setEncoder(k, v)
	}

	return
}

func (lo *PatternLayout) encode(out *bytes.Buffer, r *driver.Recorder) {
	t := r.Created
	if lo.utc {
		t = t.UTC()
	}

	// Iterate over the pieces, replacing known formats
	// Split the string into pieces by % signs
	// verbs := bytes.Split([]byte(format), []byte{'%'})

	for i, piece := range lo.verbs {
		if i == 0 && len(piece) > 0 {
			out.Write(piece)
			continue
		} else if len(piece) <= 0 {
			continue
		}
		switch piece[0] {
		case 'D':
			lo.DateEncoding(out, &t)
		case 'd':
			lo._encodeDate(out, &t)
		case 'T':
			lo.TimeEncoding(out, &t)
		case 't':
			lo._encodeTime(out, &t)
		case 'Z':
			lo.ZoneEncoding(out, &t)
		case 'L':
			lo.LevelEncoding(out, r.Level)
		case 'l':
			var b []byte
			itoa(&b, int(r.Level), -1)
			out.Write(b)
		case 'P':
			out.WriteString(r.Prefix)
		case 'S':
			lo.CallerEncoding(out, r.Source)
		case 'N':
			var b []byte
			itoa(&b, r.Line, -1)
			out.Write(b)
		case 'M':
			out.WriteString(r.Message)
		case 'F':
			lo.FieldsEncoding(out, r.Data, r.Index)
		default:
			// unknown format code. Ignored.
		}
		if len(piece) > 1 {
			out.Write(piece[1:])
		}
	}
}

// Encode log Recorder to bytes buffer.
func (lo *PatternLayout) Encode(out *bytes.Buffer, r *driver.Recorder) int {
	if r == nil {
		out.Write([]byte("<nil>"))
		return out.Len()
	} else if len(lo.verbs) == 0 {
		return out.Len()
	}

	if lo.color {
		out.Write(Encoders.Level.Begin(r.Level))
	}

	lo.encode(out, r)

	out.Write(lo.lineEnd)

	if lo.color {
		out.Write(Encoders.Level.End(r.Level))
	}
	return out.Len()
}
