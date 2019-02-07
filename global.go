// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
)

var global = New(DEBUG)

// GetLogger returns the default logger.
func GetLogger() *Logger {
	return global
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
	msg := FormatMessage(arg0, args...)
	global.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return.
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	global.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return.
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	global.intLog(CRITICAL, msg)
	return errors.New(msg)
}
