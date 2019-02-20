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
func Finest(args ...interface{}) {
	std.intLog(FINEST, args...)
}

// Fine is a wrapper for (*Logger).Fine
func Fine(args ...interface{}) {
	std.intLog(FINE, args...)
}

// Debug is a wrapper for (*Logger).Debug
func Debug(args ...interface{}) {
	std.intLog(DEBUG, args...)
}

// Trace is a wrapper for (*Logger).Trace
func Trace(args ...interface{}) {
	std.intLog(TRACE, args...)
}

// Info is a wrapper for (*Logger).Info
func Info(args ...interface{}) {
	std.intLog(INFO, args...)
}

// Warn is a wrapper for (*Logger).Warn
func Warn(args ...interface{}) error {
	msg := FormatMessage(args...)
	std.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error is a wrapper for (*Logger).Error
func Error(args ...interface{}) error {
	msg := FormatMessage(args...)
	std.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical is a wrapper for (*Logger).Critical
func Critical(args ...interface{}) error {
	msg := FormatMessage(args...)
	std.intLog(CRITICAL, msg)
	return errors.New(msg)
}
