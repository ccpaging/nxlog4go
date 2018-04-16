// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

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
	"os"
	"runtime"
	"strings"
	"time"
	"sync"
	"io"
)

// Version information
const (
	NXLOG4GO_VERSION = "nxlog4go-v0.4.4"
	NXLOG4GO_MAJOR   = 0
	NXLOG4GO_MINOR   = 4
	NXLOG4GO_BUILD   = 4
)

/****** Constants ******/

// These are the integer logging levels used by the logger
type Level int

const (
	FINEST Level = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
	_SILENT_ = 100 // SILENT is used during configuration to turn in quiet mode
)

// Logging level strings
var (
	levelStrings = [...]string{"FNST", "FINE", "DEBG", "TRAC", "INFO", "WARN", "EROR", "CRIT", "DISA"}
)

func (l Level) String() string {
	if l < 0 || int(l) > len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
}

func GetLevel(s string) (lvl Level) {
	switch s {
	case "FINEST", FINEST.String():
		lvl = FINEST
	case FINE.String():
		lvl = FINE
	case "DEBUG", DEBUG.String():
		lvl = DEBUG
	case "TRACE", TRACE.String():
		lvl = TRACE
	case "INFO", INFO.String():
		lvl = INFO
	case "WARNING", WARNING.String():
		lvl = WARNING
	case "ERROR", ERROR.String():
		lvl = ERROR
	case "CRITICAL", CRITICAL.String():
		lvl = CRITICAL
	case "DISABLE", "DISA", "SILENT", "QUIET":
		lvl = _SILENT_
	default:
		lvl = INFO
	}
	return lvl
}

/****** Variables ******/

var (
	// Default skip passed to runtime.Caller to get file name/line
	// May require tweaking if you want to wrap the logger
	LogCallerDepth = 2
	// LogBufferLength specifies how many log messages a particular log4go
	// logger can buffer at a time before writing them.
	LogBufferLength = 32
)

/****** LogRecord ******/

// A LogRecord contains all of the pertinent information for each message
type LogRecord struct {
	Level   Level     // The log level
	Created time.Time // The time at which the log message was created (nanoseconds)
	Prefix  string    // The message prefix
	Source  string    // The message source
	Line	int 	  // The source line
	Message string    // The log message
}

/****** Logger ******/

// A Logger represents an active logging object that generates lines of
// output to an io.Writer, and a collection of Filters through which 
// log messages are written. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields

	prefix string     // prefix to write at beginning of each line
	caller bool  	  // runtime caller skip

	out    io.Writer  // destination for output
	level  Level      // The log level
	layout Layout     // format record for output

	filters *Filters  // a collection of Filters
}

// NewLogger creates a new Logger. The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.
func NewLogger(out io.Writer, lvl Level, prefix string, pattern string) *Logger {
	return &Logger{
		out: out,
		level: lvl,
		caller: true,
		prefix: prefix,
		layout: NewPatternLayout(pattern),
		filters: nil,
	}
}

// New Creates a new logger with a "stderr" writer to send 
// log messages at or above lvl to standard output.
func New(lvl Level) *Logger {
	return NewLogger(os.Stderr, lvl, "", PATTERN_DEFAULT)
}

// Closes all log filters in preparation for exiting the program.
// Calling this is not really imperative, unless you want to 
// guarantee that all log messages are written.  Close removes
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

// Caller return runtime caller skip for the logger.
func (log *Logger) Caller() bool {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.caller
}

// SetCaller enable or disable the runtime caller function for the logger.
func (log *Logger) SetCaller(caller bool) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.caller = caller
	return log
}

// Level returns the output level for the logger.
func (log *Logger) Level() Level {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.level
}

// SetLevel sets the output level for the logger.
func (log *Logger) SetLevel(lvl Level) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.level = lvl
	return log
}

// Pattern returns the output PatternLayout's pattern for the logger.
func (log *Logger) Pattern() string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.layout.Get("pattern")
}

// SetPattern sets the output PatternLayout's pattern for the logger.
func (log *Logger) SetPattern(pattern string) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.layout.Set("pattern", pattern)
	return log
}

// Set option. chainable
func (log *Logger) Set(name string, v interface{}) *Logger {
	log.SetOption(name, v)
	return log
}

/*
Set option. checkable
Option names include:
	prefix  - The output prefix
	caller	- Enable or disable the runtime caller function
	level   - The output level
	pattern	- The pattern of Layout format
*/
func (log *Logger) SetOption(name string, v interface{}) error {
	switch name {
	case "prefix":
		if prefix, ok := v.(string); ok {
			log.SetPrefix(prefix)
		} else {
			return ErrBadValue
		}
	case "caller":
		caller := false
		switch value := v.(type) {
		case bool:
			caller = value
		case string:
			if strings.HasPrefix(value, "T") || strings.HasPrefix(value, "t") {
				caller = true 
			}
		default:
			return ErrBadValue
		}
		log.SetCaller(caller)
	case "level":
		lvl := INFO
		switch value := v.(type) {
		case int:
			lvl = Level(value)
		case string:
			lvl = GetLevel(value)
		default:
			return ErrBadValue
		}
		log.SetLevel(lvl)
	case "pattern":
		if pattern, ok := v.(string); ok {
			log.SetPattern(pattern)
		} else {
			return ErrBadValue
		}
	default:
		return ErrBadOption
	}
	return nil
}

// SetOutput sets the output destination for the logger.
func (log *Logger) SetOutput(w io.Writer) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.out = w
	return log
}

// Layout returns the output Layout for the logger.
func (log *Logger) Layout() Layout {
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

// Filters returns the output filters for the logger.
func (log *Logger) Filters() *Filters {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.filters
}

// SetFilters sets the output filters for the logger.
func (log *Logger) SetFilters(filters *Filters) *Logger {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.filters = filters
	return log
}

/******* Logging *******/

// Determine if any logging will be done
func (log Logger) skip(lvl Level) bool {
	if log.out != nil && lvl >= log.level {
        return false
    }

	if log.filters != nil {
		if log.filters.Skip(lvl) == false{
			return false
		}
	}
	return true
}

func intMsg(arg0 interface{}, args ...interface{}) string {
	switch first := arg0.(type) {
	case string:
		if len(args) == 0 {
			return first
		} else {
			// Use the string as a format string
			return fmt.Sprintf(first, args...)
		}
	case func() string:
		// Log the closure (no other arguments used)
		return first()
	default:
		// Build a format string so that it will be similar to Sprint
		return fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Send a log message with manual level, source, and message.
func (log Logger) Log(lvl Level, source string, arg0 interface{}, args ...interface{}) {
	if log.skip(lvl) {
		return
	}

	log.mu.Lock()
	defer log.mu.Unlock()

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Prefix:  log.prefix,
		Source:  source,
		Message: intMsg(arg0, args ...),
	}

	if log.out != nil {
    	log.out.Write(log.layout.Format(rec))
	}

	if log.filters != nil {
		log.filters.Dispatch(rec)
	}
}

// Send a log message with level, and message.
func (log Logger) intLog(lvl Level, message string) {
	log.mu.Lock()
	defer log.mu.Unlock()

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Prefix:  log.prefix,
		Message: message,
	}

	if log.caller {
		// Determine caller func - it's expensive.
		log.mu.Unlock()

		var ok bool
		_, rec.Source, rec.Line, ok = runtime.Caller(LogCallerDepth)
		if !ok {
			rec.Source = "???"
			rec.Line = 0
		}

		log.mu.Lock()
	}
	
	if log.out != nil {
    	log.out.Write(log.layout.Format(rec))
	}

	if log.filters != nil {
		log.filters.Dispatch(rec)
	}
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (log Logger) Finest(arg0 interface{}, args ...interface{}) {
	if log.skip(FINEST) {
		return
	}
	log.intLog(FINEST, intMsg(arg0, args...))
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (log Logger) Fine(arg0 interface{}, args ...interface{}) {
	if log.skip(FINE) {
		return
	}
	log.intLog(FINE, intMsg(arg0, args...))
}

// Debug is a utility method for debug log messages.
// The behavior of Debug depends on the first argument:
// - arg0 is a string
//   When given a string as the first argument, this behaves like Logf but with
//   the DEBUG log level: the first argument is interpreted as a format for the
//   latter arguments.
// - arg0 is a func()string
//   When given a closure of type func()string, this logs the string returned by
//   the closure iff it will be logged.  The closure runs at most one time.
// - arg0 is interface{}
//   When given anything else, the log message will be each of the arguments
//   formatted with %v and separated by spaces (ala Sprint).
func (log Logger) Debug(arg0 interface{}, args ...interface{}) {
	if log.skip(DEBUG) {
		return
	}
	log.intLog(DEBUG, intMsg(arg0, args...))
}

// Trace logs a message at the trace log level.
// See Debug for an explanation of the arguments.
func (log Logger) Trace(arg0 interface{}, args ...interface{}) {
	if log.skip(TRACE) {
		return
	}
	log.intLog(TRACE, intMsg(arg0, args...))
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (log Logger) Info(arg0 interface{}, args ...interface{}) {
	if log.skip(INFO) {
		return
	}
	log.intLog(INFO, intMsg(arg0, args...))
}

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See Debug for further explanation of the arguments.
func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	if !log.skip(WARNING) {
		log.intLog(WARNING, msg)
	}
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	if !log.skip(ERROR) {
		log.intLog(ERROR, msg)
	}
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (log Logger) Critical(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	if !log.skip(CRITICAL) {
		log.intLog(CRITICAL, msg)
	}
	return errors.New(msg)
}
