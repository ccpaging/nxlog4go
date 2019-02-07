// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func intMsg(arg0 interface{}, args ...interface{}) (s string) {
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
		log.Log(lvl, "", 0, intMsg(arg0, args...))
	} else {
		// Determine caller func - it's expensive.
		_, source, line, _ := runtime.Caller(LogCallerDepth)
		log.Log(lvl, source, line, intMsg(arg0, args...))
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
	msg := intMsg(arg0, args...)
	log.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	log.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (log Logger) Critical(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	log.intLog(CRITICAL, msg)
	return errors.New(msg)
}

var global = New(DEBUG)

// GetLogger returns the default logger.
func GetLogger() *Logger {
	return global
}

// Panic is compatible with `log`.
func Panic(arg0 interface{}, args ...interface{}) {
	msg := intMsg(arg0, args...)
	global.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicln is compatible with `log`.
func Panicln(arg0 interface{}, args ...interface{}) {
	msg := intMsg(arg0, args...)
	global.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicf is compatible with `log`.
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	global.intLog(CRITICAL, s)
	panic(s)
}

// Fatal is compatible with `log`.
func Fatal(arg0 interface{}, args ...interface{}) {
	global.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalln is compatible with `log`.
func Fatalln(arg0 interface{}, args ...interface{}) {
	global.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalf is compatible with `log`.
func Fatalf(format string, v ...interface{}) {
	global.intLog(ERROR, fmt.Sprintf(format, v...))
	os.Exit(0)
}

// Print is compatible with `log`.
func Print(arg0 interface{}, args ...interface{}) {
	global.intLog(INFO, arg0, args...)
}

// Println is compatible with `log`.
func Println(arg0 interface{}, args ...interface{}) {
	global.intLog(INFO, arg0, args...)
}

// Printf is compatible with `log`.
func Printf(format string, v ...interface{}) {
	global.intLog(INFO, fmt.Sprintf(format, v...))
}

// Finest log messages (see Debug() for parameter explanation).
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	global.intLog(FINEST, arg0, args...)
}

// Fine log messages (see Debug() for parameter explanation).
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	global.intLog(FINE, arg0, args...)
}

// Debug log messages.
// When given a string as the first argument, this behaves like Log
// but with the DEBUG log level (e.g. the first argument is interpreted
// as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned
// by the closure if it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments
// formatted with %v and separated by spaces (ala Sprint).
// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	global.intLog(DEBUG, arg0, args...)
}

// Trace log messages (see Debug() for parameter explanation).
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	global.intLog(TRACE, arg0, args...)
}

// Info log messages (see Debug() for parameter explanation).
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	global.intLog(INFO, arg0, args...)
}

// Warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return.
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	global.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return.
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	global.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return.
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	global.intLog(CRITICAL, msg)
	return errors.New(msg)
}
