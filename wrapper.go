// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"os"
)

var Global = New(DEBUG)

func GetLogger() *Logger {
	return Global
}

// Logs the given message and crashes the program
func Crash(arg0 interface{}, args ...interface{}) {
	panic(Global.intLog(CRITICAL, arg0, args...))
}

// Logs the given message and crashes the program
func Crashf(format interface{}, args ...interface{}) {
	panic(Global.intLog(CRITICAL, format, args...))
}

// Compatibility with `log`
func Exit(arg0 interface{}, args ...interface{}) {
	Global.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Compatibility with `log`
func Exitf(format interface{}, args ...interface{}) {
	Global.intLog(ERROR, format, args...)
	os.Exit(0)
}

// Compatibility with `log`
func Stderr(arg0 interface{}, args ...interface{}) {
	Global.intLog(ERROR, arg0, args...)
}

// Compatibility with `log`
func Stderrf(format interface{}, args ...interface{}) {
	Global.intLog(ERROR, format, args...)
}

// Compatibility with `log`
func Stdout(arg0 interface{}, args ...interface{}) {
	Global.intLog(INFO, arg0, args...)
}

// Compatibility with `log`
func Stdoutf(format interface{}, args ...interface{}) {
	Global.intLog(INFO, format, args...)
}

// Utility for finest log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	if Global.skip(FINEST) {
		return
	}
	Global.intLog(FINEST, arg0, args...)
}

// Utility for fine log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	if Global.skip(FINE) {
		return
	}
	Global.intLog(FINE, arg0, args...)
}

// Utility for debug log messages
// When given a string as the first argument, this behaves like Logf but with the DEBUG log level (e.g. the first argument is interpreted as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned by the closure iff it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments formatted with %v and separated by spaces (ala Sprint).
// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	if Global.skip(DEBUG) {
		return
	}
	Global.intLog(DEBUG, arg0, args...)
}

// Utility for trace log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	if Global.skip(TRACE) {
		return
	}
	Global.intLog(TRACE, arg0, args...)
}

// Utility for info log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	if Global.skip(INFO) {
		return
	}
	Global.intLog(INFO, arg0, args...)
}

// Utility for warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) error {
	return errors.New(Global.intLog(WARNING, arg0, args...))
}

// Utility for error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	return errors.New(Global.intLog(ERROR, arg0, args...))
}

// Utility for critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	return errors.New(Global.intLog(CRITICAL, arg0, args...))
}
