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

// Encoders includes all log recorder field encoder.
type Encoders struct {
	BeginColorizer Encoder
	EndColorizer   Encoder
	LevelEncoder   Encoder
	DateEncoder    Encoder
	TimeEncoder    Encoder
	ZoneEncoder    Encoder
	CallerEncoder  Encoder
	FieldsEncoder  Encoder
}

// DefaultEncoders allows users to configure the concrete encoders.
var DefaultEncoders Encoders = Encoders{
	BeginColorizer: NewNopEncoder(),
	EndColorizer:   NewNopEncoder(),
	LevelEncoder:   NewNopEncoder(),
	DateEncoder:    NewDateEncoder(""),
	TimeEncoder:    NewTimeEncoder(""),
	ZoneEncoder:    NewZoneEncoder(""),
	CallerEncoder:  NewCallerEncoder(""),
	FieldsEncoder:  NewFieldsEncoder(""),
}

// PatternLayout formats log Recorder.
type PatternLayout struct {
	verbs [][]byte // Split the format string into pieces by % signs

	lineEnd []byte
	utc     bool

	Encoders

	// DEPRECATED. Compatible with log4go
	_encodeDate Encoder
	_encodeTime Encoder // DEPRECATED
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
		verbs: formatToVerbs(format),

		lineEnd: DefaultLineEnd,
		utc:     false,

		Encoders: DefaultEncoders,

		// DEPRECATED. Compatible with log4go
		_encodeDate: NewDateEncoder("mdy"),
		_encodeTime: NewTimeEncoder("hhmm"),
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

func (lo *PatternLayout) setEncoder(k string, v interface{}) error {
	s, err := cast.ToString(v)
	if err != nil {
		return err
	}
	switch k {
	case "levelEncoder":
		lo.LevelEncoder = lo.LevelEncoder.Open(s)
	case "callerEncoder":
		lo.CallerEncoder = lo.CallerEncoder.Open(s)
	case "dateEncoder":
		lo.DateEncoder = lo.DateEncoder.Open(s)
	case "timeEncoder":
		lo.TimeEncoder = lo.TimeEncoder.Open(s)
	case "zoneEncoder":
		lo.ZoneEncoder = lo.ZoneEncoder.Open(s)
	case "fieldsEncoder":
		lo.FieldsEncoder = lo.FieldsEncoder.Open(s)
	default:
		return fmt.Errorf("unknown option name %s, value %#v of type %T", k, v, v)
	}

	return nil
}

// Set sets name-value option with:
//  format  - Layout format string. Auto-detecting quote string.
//  lineEnd - Line end string. Auto-detecting quote string.
//	color   - Set ansi color true, false, or "auto".
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
			if ok {
				lo.BeginColorizer = lo.BeginColorizer.Open("color")
				lo.EndColorizer = lo.EndColorizer.Open("color")
			} else {
				lo.BeginColorizer = lo.BeginColorizer.Open("nocolor")
				lo.EndColorizer = lo.EndColorizer.Open("nocolor")
			}
		} else {
			lo.BeginColorizer = lo.BeginColorizer.Open("auto")
			lo.EndColorizer = lo.EndColorizer.Open("auto")
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
	tr := &driver.Recorder{Created: t}

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
			lo.DateEncoder.Encode(out, tr)
		case 'd':
			lo._encodeDate.Encode(out, tr)
		case 'T':
			lo.TimeEncoder.Encode(out, tr)
		case 't':
			lo._encodeTime.Encode(out, tr)
		case 'Z':
			lo.ZoneEncoder.Encode(out, tr)
		case 'L':
			lo.LevelEncoder.Encode(out, r)
		case 'l':
			var b []byte
			itoa(&b, int(r.Level), -1)
			out.Write(b)
		case 'P':
			out.WriteString(r.Prefix)
		case 'S':
			lo.CallerEncoder.Encode(out, r)
		case 'N':
			var b []byte
			itoa(&b, r.Line, -1)
			out.Write(b)
		case 'M':
			out.WriteString(r.Message)
		case 'F':
			lo.FieldsEncoder.Encode(out, r)
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

	lo.BeginColorizer.Encode(out, r)
	lo.encode(out, r)
	lo.EndColorizer.Encode(out, r)

	out.Write(lo.lineEnd)
	return out.Len()
}
