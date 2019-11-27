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
//	 WARN
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
// the important ones show up as the format you want.
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
	"io"
	"os"
	"sync"
)

// Version information
const (
	Version = "nxlog4go-v1.0.2"
	Major   = 1
	Minor   = 0
	Build   = 2
)

/****** Logger ******/

// PreHook function runs before writing log record.
// Return false then skip writing log record
//
// DEPRECATED: Use appender instead.
type PreHook func(e *Entry) bool

// PostHook function runs after writing log record even BeforeLog returns false.
//
// DEPRECATED: Use appender instead.
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
	level  int       // The log level
	layout Layout    // format record to []byte for output

	preHook  PreHook
	postHook PostHook

	filters Filters // a collection of Filters
}

// NewLogger creates a new logger with a "stderr" writer to send
// formatted log messages at or above level to standard output.
func NewLogger(level int) *Logger {
	return &Logger{
		out:     os.Stderr,
		level:   level,
		caller:  true,
		prefix:  "",
		layout:  NewPatternLayout(PatternDefault),
		filters: nil,
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
func (l *Logger) Set(args ...interface{}) *Logger {
	ops, idx, _ := ArgsToMap(args)
	for _, k := range idx {
		l.SetOption(k, ops[k])
	}
	return l
}

// SetOption sets options of logger.
// Option names include:
//	prefix - The output prefix
//	caller - Enable or disable the runtime caller function
//	level  - The output level
//
// layout options...
//
// Return errors.
func (l *Logger) SetOption(k string, v interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch k {
	case "prefix":
		if prefix, err := ToString(v); err == nil {
			l.prefix = prefix
		} else {
			return err
		}
	case "caller":
		if caller, err := ToBool(v); err == nil {
			l.caller = caller
		} else {
			return err
		}
	case "level":
		switch v.(type) {
		case int:
			l.level = v.(int)
		case Level:
			l.level = int(v.(Level))
		case string:
			l.level = Level(INFO).Int(v.(string))
		default:
			return ErrBadValue
		}
	case "color":
		if color, err := ToBool(v); err == nil {
			if color {
				l.preHook = setColor
				l.postHook = resetColor
			} else {
				l.preHook = nil
				l.postHook = nil
			}
		} else {
			return err
		}
	default:
		return l.layout.SetOption(k, v)
	}
	return nil
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
