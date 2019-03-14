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
	std.Log(2, FINEST, args...)
}

// Fine is a wrapper for (*Logger).Fine
func Fine(args ...interface{}) {
	std.Log(2, FINE, args...)
}

// Debug is a wrapper for (*Logger).Debug
func Debug(args ...interface{}) {
	std.Log(2, DEBUG, args...)
}

// Trace is a wrapper for (*Logger).Trace
func Trace(args ...interface{}) {
	std.Log(2, TRACE, args...)
}

// Info is a wrapper for (*Logger).Info
func Info(args ...interface{}) {
	std.Log(2, INFO, args...)
}

// Warn is a wrapper for (*Logger).Warn
func Warn(args ...interface{}) error {
	msg := FormatMessage(args...)
	std.Log(2, WARNING, msg)
	return errors.New(msg)
}

// Error is a wrapper for (*Logger).Error
func Error(args ...interface{}) error {
	msg := FormatMessage(args...)
	std.Log(2, ERROR, msg)
	return errors.New(msg)
}

// Critical is a wrapper for (*Logger).Critical
func Critical(args ...interface{}) error {
	msg := FormatMessage(args...)
	std.Log(2, CRITICAL, msg)
	return errors.New(msg)
}
