// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"fmt"
	"strings"
)

/******* Logging *******/

// FormatMessage builds a format string by the arguments
// Return a format string which depends on the first argument:
// - No arg
//   Return ""
// - arg[0] is a string
//   When given a string as the first argument, this behaves like fmt.Sprintf
//   the first argument is interpreted as a format for the latter arguments.
// - arg[0] is a func()string
//   When given a closure of type func()string, this logs the string returned by
//   the closure if it will be logged.  The closure runs at most one time.
// - arg[0] is interface{}
//   When given anything else, the return message will be each of the arguments
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
		if len(args) == 1 {
			s = fmt.Sprint(first)
		} else {
			s = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)-1), args[1:]...)
		}
	}
	return
}

// Skip determines whether any logging will be skipped or not.
func (l *Logger) Skip(level int) bool {
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

/****** Wrapper log write functions for logger ******/

// With adds key-value pairs to the log record, note that it doesn't log until you call
// Debug, Print, Info, Warn, Error, Fatal or Panic. It only creates a log record.
func (l *Logger) With(args ...interface{}) *Entry {
	return NewEntry(l).With(args...)
}

// Log sends a log message with level and message.
// Call depth:
//  2 - Where calling the wrapper of logger.Log(...)
//  1 - Where calling logger.Log(...)
//  0 - Where internal calling entry.flush()
func (l *Logger) Log(calldepth int, level int, args ...interface{}) {
	if !l.Skip(level) {
		e := NewEntry(l)
		e.Level = level
		e.Message = FormatMessage(args...)
		e.calldepth = calldepth + 1
		e.flush()
	}
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Finest(args ...interface{}) {
	l.Log(2, FINEST, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Fine(args ...interface{}) {
	l.Log(2, FINE, args...)
}

// Debug is a utility method for debug log messages.
// See FormatMessage for an explanation of the arguments.
func (l *Logger) Debug(args ...interface{}) {
	l.Log(2, DEBUG, args...)
}

// Trace logs a message at the trace log level.
// See FormatMessage for an explanation of the arguments.
func (l *Logger) Trace(args ...interface{}) {
	l.Log(2, TRACE, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Info(args ...interface{}) {
	l.Log(2, INFO, args...)
}

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See FormatMessage for further explanation of the arguments.
func (l *Logger) Warn(args ...interface{}) error {
	msg := FormatMessage(args...)
	l.Log(2, WARN, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and FormatMessage for an explanation
// of the parameters.
func (l *Logger) Error(args ...interface{}) error {
	msg := FormatMessage(args...)
	l.Log(2, ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and FormatMessage for an explanation
// of the parameters.
func (l *Logger) Critical(args ...interface{}) error {
	msg := FormatMessage(args...)
	l.Log(2, CRITICAL, msg)
	return errors.New(msg)
}
