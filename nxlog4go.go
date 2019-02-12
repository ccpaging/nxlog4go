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
// log.Info("The time is now: %s", time.LocalTime().Format("15:04:05 MST 2006/01/02"))
//
// Usage notes:
// - The utility functions (Info, Debug, Warn, etc) derive their source from the
//   calling function, and this incurs extra overhead. It can be disabled.
// - New field prefix is adding to LogRecorder to identify different module/package
//   in large project
//
// Changes from log4go
//
// The most of interfaces and the internals have been changed have been changed, then you will
// have to update your code. Sorry! I hope it is worth.

package nxlog4go

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Version information
const (
	Version = "nxlog4go-v0.4.4"
	Major   = 0
	Minor   = 4
	Build   = 4
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
func (l Level) String() string {
	if l < 0 || int(l) >= len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
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

/****** LogRecord ******/

// A LogRecord contains all of the pertinent information for each message
type LogRecord struct {
	Level   Level     // The log level
	Created time.Time // The time at which the log message was created (nanoseconds)
	Prefix  string    // The message prefix
	Source  string    // The message source
	Line    int       // The source line
	Message string    // The log message
}

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

	out    io.Writer // destination for output
	level  Level     // The log level
	layout Layout    // format record for output

	filters Filters // a collection of Filters
}

// NewLogger creates a new logger with a "stderr" writer to send
// formatted log messages at or above lvl to standard output.
func NewLogger(lvl Level) *Logger {
	return &Logger{
		out:     os.Stderr,
		level:   lvl,
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
func (log *Logger) Shutdown() {
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.filters != nil {
		log.filters.Close()
		log.filters = nil
	}
}

// Prefix returns the output prefix for the logger.
func (log *Logger) Prefix() string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.prefix
}

// SetPrefix sets the output prefix for the logger.
func (log *Logger) SetPrefix(prefix string) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.prefix = prefix
	return log
}

// Set option. chainable
func (log *Logger) Set(k string, v interface{}) *Logger {
	log.SetOption(k, v)
	return log
}

// SetOption sets options of logger.
// Option names include:
//	prefix  - The output prefix
//	caller	- Enable or disable the runtime caller function
//	level   - The output level
//	pattern	- The pattern of Layout format
// Return errors.
func (log *Logger) SetOption(k string, v interface{}) (err error) {
	err = nil

	switch k {
	case "prefix":
		prefix := ""
		if prefix, err = ToString(v); err == nil {
			log.SetPrefix(prefix)
		}
	case "caller":
		caller := false
		if caller, err = ToBool(v); err == nil {
			log.mu.Lock()
			log.caller = caller
			log.mu.Unlock()
		}
	case "level":
		log.mu.Lock()
		defer log.mu.Unlock()
		switch v.(type) {
		case int:
			log.level = Level(v.(int))
		case Level:
			log.level = v.(Level)
		case string:
			log.level = GetLevel(v.(string))
		default:
			err = ErrBadValue
		}
	default:
		return log.layout.SetOption(k, v)
	}
	return
}

// SetOutput sets the output destination for the logger.
func (log *Logger) SetOutput(w io.Writer) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.out = w
	return log
}

// GetLayout returns the output layout for the logger.
func (log *Logger) GetLayout() Layout {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.layout
}

// SetLayout sets the output layout for the logger.
func (log *Logger) SetLayout(layout Layout) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.layout = layout
	return log
}

// GetFilters returns the output filters for the logger.
func (log *Logger) GetFilters() Filters {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.filters
}

// SetFilters sets the output filters for the logger.
func (log *Logger) SetFilters(filters Filters) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.filters = filters
	return log
}

/******* Logging *******/

// FormatMessage builds a format string by the arguments
// Return a format string
func FormatMessage(arg0 interface{}, args ...interface{}) (s string) {
	switch first := arg0.(type) {
	case string:
		if len(args) == 0 {
			s = first
		} else {
			// Use the string as a format string
			s = fmt.Sprintf(first, args...)
		}
	case func() string:
		// Log the closure (no other arguments used)
		s = first()
	default:
		// Build a format string so that it will be similar to Sprint
		s = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	return
}

// Determine if any logging will be done.
func (log Logger) skip(lvl Level) bool {
	if log.out != nil && lvl >= log.level {
		return false
	}

	if log.filters != nil {
		if log.filters.Skip(lvl) == false {
			return false
		}
	}

	// log.out == nil and log.filters == nil
	// or lvl < log.Level
	return true
}

func (log Logger) withoutLock(calldepth int, lvl Level, message string) {
	source, line := "", 0
	if log.caller {
		log.mu.Unlock()
		// Determine caller func - it's expensive.
		_, source, line, _ = runtime.Caller(calldepth)
		log.mu.Lock()
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Prefix:  log.prefix,
		Source:  source,
		Line:    line,
		Message: message,
	}

	if log.out != nil && lvl >= log.level {
		log.out.Write(log.layout.Format(rec))
	}

	if log.filters != nil {
		log.filters.Dispatch(rec)
	}
}

// Send a log message with level, and message.
func (log Logger) intLog(lvl Level, arg0 interface{}, args ...interface{}) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.skip(lvl) {
		return
	}

	log.withoutLock(2, lvl, FormatMessage(arg0, args...))
}

// Log sends a log message with calldepth, level, and message.
func (log Logger) Log(calldepth int, lvl Level, message string) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.skip(lvl) {
		return
	}

	log.withoutLock(calldepth+1, lvl, message)
}
