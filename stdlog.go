// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"io"
)

var std = NewLogger(INFO)

// GetLogger returns the default logger.
func GetLogger() *Logger {
	return std
}

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) *Logger {
	return std.SetOutput(w)
}

// Finest is a wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	std.Log(2, FINEST, arg0, args...)
}

// Fine is a wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	std.Log(2, FINE, arg0, args...)
}

// Debug is a wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	std.Log(2, DEBUG, arg0, args...)
}

// Trace is a wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	std.Log(2, TRACE, arg0, args...)
}

// Info is a wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	std.Log(2, INFO, arg0, args...)
}

// Warn is a wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) error {
	msg := ArgsToString(arg0, args...)
	std.Log(2, WARN, msg)
	return errors.New(msg)
}

// Error is a wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	msg := ArgsToString(arg0, args...)
	std.Log(2, ERROR, msg)
	return errors.New(msg)
}

// Critical is a wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	msg := ArgsToString(arg0, args...)
	std.Log(2, CRITICAL, msg)
	return errors.New(msg)
}
