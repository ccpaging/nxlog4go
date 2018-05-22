// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"os"
)

var global = New(DEBUG)

// Return the default logger 
func GetLogger() *Logger {
	return global
}

// Compatibility with `log`
func Panic(arg0 interface{}, args ...interface{}) {
	msg := intMsg(arg0, args...)
	global.intLog(CRITICAL, msg)
	panic(msg)
}

// Compatibility with `log`
func Panicf(format interface{}, args ...interface{}) {
	msg := intMsg(format, args...)
	global.intLog(CRITICAL, msg)
	panic(msg)
}

// Compatibility with `log`
func Fatal(arg0 interface{}, args ...interface{}) {
	global.intLog(ERROR, intMsg(arg0, args...))
	os.Exit(0)
}

// Compatibility with `log`
func Fatalf(format interface{}, args ...interface{}) {
	global.intLog(ERROR, intMsg(format, args...))
	os.Exit(0)
}

// Compatibility with `log`
func Print(arg0 interface{}, args ...interface{}) {
	global.intLog(INFO, intMsg(arg0, args...))
}

// Compatibility with `log`
func Printf(format interface{}, args ...interface{}) {
	global.intLog(INFO, intMsg(format, args...))
}

// Utility for finest log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	if global.skip(FINEST) {
		return
	}
	global.intLog(FINEST, intMsg(arg0, args...))
}

// Utility for fine log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	if global.skip(FINE) {
		return
	}
	global.intLog(FINE, intMsg(arg0, args...))
}

// Utility for debug log messages
// When given a string as the first argument, this behaves like Logf but with the DEBUG log level (e.g. the first argument is interpreted as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned by the closure iff it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments formatted with %v and separated by spaces (ala Sprint).
// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	if global.skip(DEBUG) {
		return
	}
	global.intLog(DEBUG, intMsg(arg0, args...))
}

// Utility for trace log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	if global.skip(TRACE) {
		return
	}
	global.intLog(TRACE, intMsg(arg0, args...))
}

// Utility for info log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	if global.skip(INFO) {
		return
	}
	global.intLog(INFO, intMsg(arg0, args...))
}

// Utility for warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	if !global.skip(WARNING) {
		global.intLog(WARNING, msg)
	}
	return errors.New(msg)
}

// Utility for error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	if !global.skip(ERROR) {
		global.intLog(ERROR, msg)
	}
	return errors.New(msg)
}

// Utility for critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	msg := intMsg(arg0, args...)
	if !global.skip(CRITICAL) {
		global.intLog(CRITICAL, msg)
	}
	return errors.New(msg)
}
