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
	"sync"

	"github.com/ccpaging/nxlog4go/cast"
	"github.com/ccpaging/nxlog4go/driver"
)

// Version information
const (
	Version = "nxlog4go-v2.0.2"
	Major   = 2
	Minor   = 0
	Build   = 2
)

/****** Logger ******/

// A Logger represents an active logging object that generates lines of
// output to an io.Writer, and a collection of Filters through which
// log messages are written. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu sync.Mutex // ensures atomic writes; protects the following fields

	prefix string // prefix to write at beginning of each line
	caller bool   // runtime caller skip

	*stdf

	filters map[string]*driver.Filter // a collection of Filter
}

const (
	stdfName string = "stdout"
)

// NewLogger creates a new logger with a "stderr" writer to send
// formatted log messages at or above level to standard output.
func NewLogger(level int) *Logger {
	l := &Logger{
		caller:  true,
		filters: make(map[string]*driver.Filter),
	}
	l.stdf = newStdf(level)
	l.filters[stdfName] = l.stdf.Filter
	return l
}

// SetOptions sets name-value pair options.
//
// Return *Logger.
func (l *Logger) SetOptions(args ...interface{}) *Logger {
	ops, idx, _ := driver.ArgsToMap(args)
	for _, k := range idx {
		l.Set(k, ops[k])
	}
	return l
}

// Set sets name-value option with:
//	prefix - The output prefix
//	caller - Enable or disable the runtime caller function
//	level  - The output level
//
// layout options...
//
// Return errors.
func (l *Logger) Set(k string, v interface{}) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var (
		s  string
		ok bool
		n  int
	)

	switch k {
	case "prefix":
		if s, err = cast.ToString(v); err == nil {
			l.prefix = s
		}
	case "caller":
		if ok, err = cast.ToBool(v); err == nil {
			l.caller = ok
		}
	case "level":
		if n, err = Level(INFO).IntE(v); err == nil {
			l.stdf.setLevel(n)
		}
	default:
		return l.stdf.Layout.Set(k, v)
	}

	return
}

// Layout returns the output layout for the logger.
func (l *Logger) Layout() driver.Layout {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stdf.Layout
}

// SetLayout sets the output layout for the logger.
func (l *Logger) SetLayout(layout driver.Layout) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stdf.Layout = layout
	return l
}

// Enable sets the standard filter's Enabler to deny all,
// or restores the default at/above level enabler.
func (l *Logger) Enable(enb bool) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stdf.enable(enb)
	return l
}

// Filters returns the output filters for the logger.
func (l *Logger) Filters() map[string]*driver.Filter {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.filters
}

// Dispatch encodes a log recorder to bytes and writes it.
func (l *Logger) Dispatch(r *driver.Recorder) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range l.filters {
		if f != nil {
			f.Dispatch(r)
		}
	}
}

// Close closes all log filters in preparation for exiting the program.
// Calling this is not really imperative, unless you want to
// guarantee that all log messages are written.
//
// Notice: Close() removes all filters (and thus all appenders) except "stdout"
// from the logger.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for name, f := range l.filters {
		if name == stdfName {
			continue
		}
		if f != nil {
			f.Close()
		}
		delete(l.filters, name)
	}
}

func (l *Logger) enabled(level int) bool {
	if l.out != nil && level >= l.level {
		return true
	}

	if l.filters != nil {
		return true
	}

	// l.out == nil and l.filters == nil
	// or level < l.Level
	return false
}
