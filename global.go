// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
)

var std = NewLogger(INFO)

// GetLogger returns the default logger.
func GetLogger() *Logger {
	return std
}

// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	std.intLog(FINEST, arg0, args...)
}

// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	std.intLog(FINE, arg0, args...)
}

// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	std.intLog(DEBUG, arg0, args...)
}

// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	std.intLog(TRACE, arg0, args...)
}

// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	std.intLog(INFO, arg0, args...)
}

// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	std.intLog(WARNING, msg)
	return errors.New(msg)
}

// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	std.intLog(ERROR, msg)
	return errors.New(msg)
}

// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	std.intLog(CRITICAL, msg)
	return errors.New(msg)
}