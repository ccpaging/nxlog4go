// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strconv"
)

// Layout is is an interface for formatting log record
type Layout interface {
	// Set option about the Layout. The options should be set as default.
	// Chainable.
	Set(args ...interface{}) Layout

	// Set option about the Layout. The options should be set as default.
	// Checkable
	SetOption(name string, v interface{}) error

	// This will be called to log a Entry message.
	Encode(out *bytes.Buffer, e *Entry) int
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
func NewPatternLayout(format string, args ...interface{}) Layout {
	lo := &PatternLayout{
		verbs:   formatToVerbs(format),
		lineEnd: DefaultLineEnd,

		utc: false,

		EncodeLevel:  NewLevelEncoder("std"),
		EncodeCaller: shortCallerEncoder,
		EncodeDate:   NewDateEncoder("cymdSlash"),
		EncodeTime:   NewTimeEncoder("hms"),
		EncodeZone:   NewZoneEncoder("mst"),
		EncodeFields: keyvalFieldsEncoder,

		// DEPRECATED. Compatible with log4go
		_encodeDate: NewDateEncoder("mdy"),
		_encodeTime: NewTimeEncoder("hhmm"),
	}
	return lo.Set(args...)
}

// NewJSONLayout creates a new layout which encode log record as JSON format.
func NewJSONLayout(args ...interface{}) Layout {
	jsonFormat := "{\"Level\":%l,\"Created\":\"%T\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%N,\"Message\":\"%M\"%F}"
	lo := NewPatternLayout(jsonFormat, args...)
	return lo.Set("timeEncoder", "rfc3339nano", "fieldsEncoder", "json")
}

// NewCSVLayout creates a new layout which encode log record as CSV format.
func NewCSVLayout(args ...interface{}) Layout {
	csvFormat := "%D|%T|%L|%P|%S:%N|%M%F"
	lo := NewPatternLayout(csvFormat, args...)
	return lo.Set("fieldsEncoder", "csv")
}

// Set options of layout. chainable
func (lo *PatternLayout) Set(args ...interface{}) Layout {
	ops, idx, _ := ArgsToMap(args)
	for _, k := range idx {
		lo.SetOption(k, ops[k])
	}
	return lo
}

func (lo *PatternLayout) setEncoder(k string, v interface{}) (err error) {
	switch k {
	case "levelEncoder":
		if _, ok := v.(string); ok {
			lo.EncodeLevel = NewLevelEncoder(v.(string))
		} else {
			err = ErrBadValue
		}
	case "callerEncoder":
		if _, ok := v.(string); ok {
			lo.EncodeCaller.SetAs(v.(string))
		} else {
			err = ErrBadValue
		}
	case "dateEncoder":
		if _, ok := v.(string); ok {
			lo.EncodeDate = NewDateEncoder(v.(string))
		} else {
			err = ErrBadValue
		}
	case "timeEncoder":
		if _, ok := v.(string); ok {
			lo.EncodeTime = NewTimeEncoder(v.(string))
		} else {
			err = ErrBadValue
		}
	case "zoneEncoder":
		if _, ok := v.(string); ok {
			lo.EncodeZone = NewZoneEncoder(v.(string))
		} else {
			err = ErrBadValue
		}
	case "fieldsEncoder":
		if _, ok := v.(string); ok {
			lo.EncodeFields.SetAs(v.(string))
		} else {
			err = ErrBadValue
		}
	default:
		return ErrBadOption
	}

	return
}

// SetOption set option with:
//  format	- Layout format string. Auto-detecting quote string.
//  lineEnd - line end string. Auto-detecting quote string.
//	utc     - Log record time zone: local or utc.
//
// Known encoder types are (The option's name and value are case-sensitive):
//  levelEncoder  - "upper", "upperColor", "lower", "lowerColor", "std" is default.
//	callerEncoder - "nopath", "fullpath", "shortpath" is default.
//  dateEncoder   - "dmy", "mdy", "cymdDash", "cymdDot", "cymdSlash" is default.
//  timeEncoder   - "hhmm",  "hms.us", "iso8601", "rfc3339nano", "hms" is default.
//  zoneEncoder   - "rfc3339", "iso8601", "mst" is default.
//  fieldsEncoder - "csv", "json", "keyval" is default.
//
// Known format codes are:
//	%D - Date (2006/01/02)
//	%T - Time (15:04:05)
//	%Z - Zone (-0700)
//	%L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
//	%l - Integer level
//	%P - Prefix
//	%S - Source
//	%N - Line number
//	%M - Message
//  %F - Data fields in "key=value" format
//
// DEPRECATED:
//	%d - Date (01/02/06). Note: Do not using with %D in a layout
//		 Replacing with setting "dateEncoder" as "mdy".
//	%t - Time (15:04). Note: Do not using with %T in a layout
//		 Replacing with setting "timeEncoder" as "hhmm".
//
//	Ignores other unknown format codes
func (lo *PatternLayout) SetOption(k string, v interface{}) (err error) {
	switch k {
	case "format", "pattern":
		if format, err := ToString(v); err == nil && len(format) > 0 {
			lo.verbs = formatToVerbs(format)
		}
	case "lineEnd":
		if lineEnd, err := ToString(v); err == nil {
			if unq, err := strconv.Unquote(lineEnd); err == nil {
				lineEnd = unq
			}
			lo.lineEnd = []byte(lineEnd)
		}
	case "color":
		if color, err := ToBool(v); err == nil {
			lo.color = color
		}
	case "utc":
		if utc, err := ToBool(v); err == nil {
			lo.utc = utc
		}
	default:
		return lo.setEncoder(k, v)
	}

	return
}

// Encode Entry to out buffer.
// Return len.
func (lo *PatternLayout) Encode(out *bytes.Buffer, e *Entry) int {
	if e == nil {
		out.Write([]byte("<nil>"))
		return out.Len()
	} else if len(lo.verbs) == 0 {
		return out.Len()
	}

	t := e.Created
	if lo.utc {
		t = t.UTC()
	}

	if lo.color {
		out.Write(Level(e.Level).Color())
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
			lo.EncodeLevel(out, e.Level)
		case 'l':
			var b []byte
			itoa(&b, int(e.Level), -1)
			out.Write(b)
		case 'P':
			out.WriteString(e.Prefix)
		case 'S':
			lo.EncodeCaller(out, e.Source)
		case 'N':
			var b []byte
			itoa(&b, e.Line, -1)
			out.Write(b)
		case 'M':
			out.WriteString(e.Message)
		case 'F':
			lo.EncodeFields(out, e.Data, e.index)
		default:
			// unknown format code. Ignored.
		}
		if len(piece) > 1 {
			out.Write(piece[1:])
		}
	}

	out.Write(lo.lineEnd)
	if lo.color {
		out.Write(ResetColor.Bytes())
	}
	return out.Len()
}
