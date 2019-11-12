// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.
//
// Package nxlog4go provides simple, fast, low cost, and extensible logging.
// It can be used in test, development, and production environment.
//
// Logger
//
// - prefix, to write at beginning of each line
// - Enable / disable caller func since it's expensive.
// - The default console output compatibility with go log
//	 level, the log level
// 	 out, io.Writer / io.MultiWriter
//   layout, specifies how the data will be written
// - filters, point to filters map
//
// Filters
//
// - The filters map indexed with tag name
//
// Filter
//
// - level, the log level
// 	 FINEST
//	 FINE
//	 DEBUG
//	 TRACE
//	 INFO
//	 WARNING
//	 ERROR
//	 CRITICAL
// - appender
//
// Appender
//
// - An interface for anything
// - Write function should be able to write logs
// - Close function clean up anything lingering about the Appender
// - SetOption function. Configurable
// - Extensible. Anyone can port own appender as part of nxlog4go.
//
// Layout
//
// - With time stamp cached
// - fast byte convert
// - The default PatternLayout is easy to use.
//
// Enhanced Logging
//
// This is inspired by the logging functionality in log4go. Essentially, you create a Logger
// object with a console writer or create output filters for it. You can send whatever you
// want to the Logger, and it will filter and formatter that based on your settings and
// send it to the outputs. This way, you can put as much debug code in your program as
// you want, and when you're done you can filter out the mundane messages so only
// the important ones show up as the pattern you want.
//
// Utility functions are provided to make life easier.
//
// You may using your configuration file format as same as your project's.
//
// You may extend your own appender for your needs.
//
// Here is some example code to get started:
//
// log := nxlog4go.New(nxlog4go.DEBUG)
// l.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
//
// Usage notes:
// - The utility functions (Info, Debug, Warn, etc) derive their source from the
//   calling function, and this incurs extra overhead. It can be disabled.
// - Adding new Entry field "prefix" to identify different module/package
//   in large project
//
// Changes from log4go
//
// The most of interfaces and the internals have been changed have been changed, then you will
// have to update your code. Sorry! I hope it is worth.

package nxlog4go

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync"
)

// Version information
const (
	Version = "nxlog4go-v1.0.1"
	Major   = 0
	Minor   = 9
	Build   = 5
)

/****** Constants ******/

// Level is the integer logging levels
type Level int

// logging levels used by the logger
const (
	FINEST Level = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
	SILENT = 100 // SILENT is used during configuration to turn in quiet mode
)

// Logging level strings
var (
	levelStrings = [...]string{"FNST", "FINE", "DEBG", "TRAC", "INFO", "WARN", "EROR", "CRIT"}
)

// String return the string of integer Level
func (level Level) String() string {
	if level < 0 || int(level) >= len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(level)]
}

// GetLevel return the integer level of string
func GetLevel(s string) (level Level) {
	switch strings.ToUpper(s) {
	case "FINEST", FINEST.String():
		level = FINEST
	case FINE.String():
		level = FINE
	case "DEBUG", DEBUG.String():
		level = DEBUG
	case "TRACE", TRACE.String():
		level = TRACE
	case "INFO", INFO.String():
		level = INFO
	case "WARNING", WARNING.String():
		level = WARNING
	case "ERROR", ERROR.String():
		level = ERROR
	case "CRITICAL", CRITICAL.String():
		level = CRITICAL
	case "DISABLE", "DISA", "SILENT", "QUIET":
		level = SILENT
	default:
		level = INFO
	}
	return level
}

/****** Variables ******/

var (
	// LogBufferLength specifies how many log messages a particular log4go
	// logger can buffer at a time before writing them.
	LogBufferLength = 32
)

/****** Errors ******/

var (
	// ErrBadOption is the errors of bad option
	ErrBadOption = errors.New("Invalid or unsupported option")
	// ErrBadValue is the errors of bad value
	ErrBadValue = errors.New("Invalid option value")
)

/****** Logger ******/

// PreHook function runs before writing log record.
// Return false then skip writing log record
type PreHook func(e *Entry) bool

// PostHook function runs after writing log record even BeforeLog returns false.
type PostHook func(e *Entry, n int, err error)

// A Logger represents an active logging object that generates lines of
// output to an io.Writer, and a collection of Filters through which
// log messages are written. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu sync.Mutex // ensures atomic writes; protects the following fields

	prefix string // prefix to write at beginning of each line
	flag   int    // properties compatible with go std log
	caller bool   // runtime caller skip

	out    io.Writer // destination for output
	level  Level     // The log level
	layout Layout    // format record for output

	preHook  PreHook
	postHook PostHook

	filters Filters // a collection of Filters
}

// NewLogger creates a new logger with a "stderr" writer to send
// formatted log messages at or above level to standard output.
func NewLogger(level Level) *Logger {
	return &Logger{
		out:     os.Stderr,
		level:   level,
		caller:  true,
		prefix:  "",
		layout:  NewPatternLayout(PatternDefault),
		filters: nil,
	}
}

// Shutdown closes all log filters in preparation for exiting the program.
// Calling this is not really imperative, unless you want to
// guarantee that all log messages are written.  Close() removes
// all filters (and thus all appenders) from the logger.
func (l *Logger) Shutdown() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.filters != nil {
		l.filters.Close()
		l.filters = nil
	}
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// SetPrefix sets the output prefix for the logger.
func (l *Logger) SetPrefix(prefix string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
	return l
}

// Set option. chainable
func (l *Logger) Set(k string, v interface{}) *Logger {
	l.SetOption(k, v)
	return l
}

// SetOption sets options of logger.
// Option names include:
//	prefix  - The output prefix
//	caller	- Enable or disable the runtime caller function
//	level   - The output level
//	pattern	- The pattern of Layout format
// Return errors.
func (l *Logger) SetOption(k string, v interface{}) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	err = nil

	switch k {
	case "prefix":
		prefix := ""
		if prefix, err = ToString(v); err == nil {
			l.prefix = prefix
		}
	case "caller":
		caller := false
		if caller, err = ToBool(v); err == nil {
			l.caller = caller
		}
	case "level":
		switch v.(type) {
		case int:
			l.level = Level(v.(int))
		case Level:
			l.level = v.(Level)
		case string:
			l.level = GetLevel(v.(string))
		default:
			err = ErrBadValue
		}
	case "color":
		color := false
		color, err = ToBool(v)
		if color {
			l.preHook = setColor
			l.postHook = resetColor
		} else {
			l.preHook = nil
			l.postHook = nil
		}
	default:
		return l.layout.SetOption(k, v)
	}
	return
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	return l
}

// Layout returns the output layout for the logger.
func (l *Logger) Layout() Layout {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.layout
}

// SetLayout sets the output layout for the logger.
func (l *Logger) SetLayout(layout Layout) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.layout = layout
	return l
}

// Filters returns the output filters for the logger.
func (l *Logger) Filters() Filters {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.filters
}

// SetFilters sets the output filters for the logger.
func (l *Logger) SetFilters(filters Filters) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filters = filters
	return l
}
