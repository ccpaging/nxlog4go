// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/ccpaging/nxlog4go/cast"
	"github.com/ccpaging/nxlog4go/driver"
)

// Layout is is an interface for formatting log record
type Layout interface {
	// SetOption sets option about the Layout. Checkable.
	Set(name string, v interface{}) error

	// Encode will be called to encode a log recorder to bytes.
	Encode(out *bytes.Buffer, r *driver.Recorder) int
}

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
	// FormatLogLog is format for internal log
	FormatLogLog = "%T %P %L %M"
)

// PatternLayout formats log record
type PatternLayout struct {
	verbs   [][]byte // Split the format string into pieces by % signs
	lineEnd []byte

	utc   bool
	color bool

	EncodeLevel  LevelEncoder
	EncodeCaller CallerEncoder
	EncodeDate   DateEncoder
	EncodeTime   TimeEncoder
	EncodeZone   ZoneEncoder
	EncodeFields FieldsEncoder

	// DEPRECATED. Compatible with log4go
	_encodeDate DateEncoder
	_encodeTime TimeEncoder // DEPRECATED
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

// NewPatternLayout creates a new layout which encode log record to bytes.
func NewPatternLayout(format string, args ...interface{}) *PatternLayout {
	lo := &PatternLayout{
		verbs:   formatToVerbs(format),
		lineEnd: DefaultLineEnd,

		utc: false,

		EncodeLevel:  NewLevelEncoder("std"),
		EncodeCaller: NewCallerEncoder("shortpath"),
		EncodeDate:   NewDateEncoder("cymdSlash"),
		EncodeTime:   NewTimeEncoder("hms"),
		EncodeZone:   NewZoneEncoder("mst"),
		EncodeFields: NewFieldsEncoder("std"),

		// DEPRECATED. Compatible with log4go
		_encodeDate: NewDateEncoder("mdy"),
		_encodeTime: NewTimeEncoder("hhmm"),
	}
	lo.SetOptions(args...)
	return lo
}

// NewJSONLayout creates a new layout which encode log record as JSON format.
func NewJSONLayout(args ...interface{}) Layout {
	jsonFormat := "{\"Level\":%l,\"Created\":\"%T\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%N,\"Message\":\"%M\"%F}"
	lo := NewPatternLayout(jsonFormat, args...)
	lo.SetOptions("timeEncoder", "rfc3339nano", "fieldsEncoder", "json")
	return lo
}

// NewCSVLayout creates a new layout which encode log record as CSV format.
func NewCSVLayout(args ...interface{}) Layout {
	csvFormat := "%D|%T|%L|%P|%S:%N|%M%F"
	lo := NewPatternLayout(csvFormat, args...)
	lo.SetOptions("fieldsEncoder", "csv")
	return lo
}

// SetOptions sets name-value pair options.
//
// Return Layout interface.
func (lo *PatternLayout) SetOptions(args ...interface{}) Layout {
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
			lo.EncodeLevel = NewLevelEncoder(s)
		}
	case "callerEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.EncodeCaller = NewCallerEncoder(s)
		}
	case "dateEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.EncodeDate = NewDateEncoder(s)
		}
	case "timeEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.EncodeTime = NewTimeEncoder(s)
		}
	case "zoneEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.EncodeZone = NewZoneEncoder(s)
		}
	case "fieldsEncoder":
		if s, err = cast.ToString(v); err == nil {
			lo.EncodeFields = NewFieldsEncoder(s)
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
//  %d - Date (01/02/06). Note: Do not using with %D in a layout
//       Replacing with setting "dateEncoder" as "mdy".
//  %t - Time (15:04). Note: Do not using with %T in a layout
//       Replacing with setting "timeEncoder" as "hhmm".
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
			lo.EncodeDate(out, &t)
		case 'd':
			lo._encodeDate(out, &t)
		case 'T':
			lo.EncodeTime(out, &t)
		case 't':
			lo._encodeTime(out, &t)
		case 'Z':
			lo.EncodeZone(out, &t)
		case 'L':
			lo.EncodeLevel(out, r.Level)
		case 'l':
			var b []byte
			itoa(&b, int(r.Level), -1)
			out.Write(b)
		case 'P':
			out.WriteString(r.Prefix)
		case 'S':
			lo.EncodeCaller(out, r.Source)
		case 'N':
			var b []byte
			itoa(&b, r.Line, -1)
			out.Write(b)
		case 'M':
			out.WriteString(r.Message)
		case 'F':
			lo.EncodeFields(out, r.Data, r.Index)
		default:
			// unknown format code. Ignored.
		}
		if len(piece) > 1 {
			out.Write(piece[1:])
		}
	}
}

// Encode Entry to out buffer.
// Return len.
func (lo *PatternLayout) Encode(out *bytes.Buffer, r *driver.Recorder) int {
	if r == nil {
		out.Write([]byte("<nil>"))
		return out.Len()
	} else if len(lo.verbs) == 0 {
		return out.Len()
	}

	if lo.color {
		out.Write(Level(r.Level).ColorBytes())
	}

	lo.encode(out, r)

	out.Write(lo.lineEnd)

	if lo.color {
		out.Write(Level(r.Level).ColorReset())
	}
	return out.Len()
}
