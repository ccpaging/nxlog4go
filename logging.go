// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

/******* Logging *******/

// FormatMessage builds a format string by the arguments
// Return a format string which depends on the first argument:
// - No arg
//   Return ""
// - arg[0] is a string
//   When given a string as the first argument, this behaves like Logf but with
//   the DEBUG log level: the first argument is interpreted as a format for the
//   latter arguments.
// - arg[0] is a func()string
//   When given a closure of type func()string, this logs the string returned by
//   the closure if it will be logged.  The closure runs at most one time.
// - arg[0] is interface{}
//   When given anything else, the log message will be each of the arguments
//   formatted with %v and separated by spaces (ala Sprint).
func FormatMessage(args ...interface{}) (s string) {
	if len(args) == 0 {
		return ""
	}
	switch first := args[0].(type) {
	case string:
		if len(args) == 1 {
			s = first
		} else {
			// Use the string as a format string
			s = fmt.Sprintf(first, args[1:]...)
		}
	case func() string:
		// Log the closure (no other arguments used)
		s = first()
	default:
		// Build a format string so that it will be similar to Sprint
		s = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)-1), args[1:]...)
	}
	return
}

// Skip determines whether any logging will be skipped or not.
func (l *Logger) Skip(level Level) bool {
	if l.out != nil && level >= l.level {
		return false
	}

	if l.filters != nil {
		if l.filters.Skip(level) == false {
			return false
		}
	}

	// l.out == nil and l.filters == nil
	// or level < l.Level
	return true
}

func (l *Logger) write(calldepth int, lvl Level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	source, line := "", 0
	if l.caller {
		l.mu.Unlock()
		// Determine caller func - it's expensive.
		_, source, line, _ = runtime.Caller(calldepth)
		l.mu.Lock()
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Prefix:  l.prefix,
		Source:  source,
		Line:    line,
		Message: message,
	}

	result := true
	if l.preHook != nil {
		result = l.preHook(l.out, rec)
	}
	var (
		n   int
		err error
	)
	if result && l.out != nil && lvl >= l.level {
		l.out.Write(l.layout.Format(rec))
	}
	if l.postHook != nil {
		l.postHook(l.out, rec, n, err)
	}

	if l.filters != nil {
		l.filters.Dispatch(rec)
	}
}

// Send a log message with level, and message.
func (l *Logger) intLog(lvl Level, args ...interface{}) {
	if l.Skip(lvl) {
		return
	}

	l.write(3, lvl, FormatMessage(args...))
}

// Log sends a log message with calldepth, level, and message.
func (l *Logger) Log(level Level, args ...interface{}) {
	if l.Skip(level) {
		return
	}

	l.write(2, level, FormatMessage(args...))
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Finest(args ...interface{}) {
	l.intLog(FINEST, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Fine(args ...interface{}) {
	l.intLog(FINE, args...)
}

// Debug is a utility method for debug log messages.
// See FormatMessage for an explanation of the arguments.
func (l *Logger) Debug(args ...interface{}) {
	l.intLog(DEBUG, args...)
}

// Trace logs a message at the trace log level.
// See FormatMessage for an explanation of the arguments.
func (l *Logger) Trace(args ...interface{}) {
	l.intLog(TRACE, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Info(args ...interface{}) {
	l.intLog(INFO, args...)
}

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See FormatMessage for further explanation of the arguments.
func (l *Logger) Warn(args ...interface{}) error {
	msg := FormatMessage(args...)
	l.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and FormatMessage for an explanation
// of the parameters.
func (l *Logger) Error(args ...interface{}) error {
	msg := FormatMessage(args...)
	l.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and FormatMessage for an explanation
// of the parameters.
func (l *Logger) Critical(args ...interface{}) error {
	msg := FormatMessage(args...)
	l.intLog(CRITICAL, msg)
	return errors.New(msg)
}
