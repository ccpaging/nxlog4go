// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

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

// Send a log message with level, and message.
func (log Logger) intLog(lvl Level, arg0 interface{}, args ...interface{}) {
	if log.skip(lvl) {
		return
	}
	if !log.caller {
		log.Log(lvl, "", 0, FormatMessage(arg0, args...))
	} else {
		// Determine caller func - it's expensive.
		_, source, line, _ := runtime.Caller(LogCallerDepth)
		log.Log(lvl, source, line, FormatMessage(arg0, args...))
	}
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (log Logger) Finest(arg0 interface{}, args ...interface{}) {
	log.intLog(FINEST, arg0, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (log Logger) Fine(arg0 interface{}, args ...interface{}) {
	log.intLog(FINE, arg0, args...)
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
	log.intLog(DEBUG, arg0, args...)
}

// Trace logs a message at the trace log level.
// See Debug for an explanation of the arguments.
func (log Logger) Trace(arg0 interface{}, args ...interface{}) {
	log.intLog(TRACE, arg0, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (log Logger) Info(arg0 interface{}, args ...interface{}) {
	log.intLog(INFO, arg0, args...)
}

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See Debug for further explanation of the arguments.
func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	log.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	log.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (log Logger) Critical(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	log.intLog(CRITICAL, msg)
	return errors.New(msg)
}
