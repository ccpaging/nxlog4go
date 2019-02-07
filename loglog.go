// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
)

var loglog *Logger

// GetLogLog creates internal logger if not existed, and returns it.
// This logger used to output log statements from within the package.
// Do not set any filters.
func GetLogLog() *Logger {
	if loglog == nil {
		loglog = New(DEBUG).Set("prefix", "lg4g").Set("pattern", "%T %P %L %M\n").Set("caller", false)
	}
	return loglog
}

// LogLogDebug logs a message at the debug log level.
func LogLogDebug(arg0 interface{}, args ...interface{}) {
	if loglog != nil {
		loglog.intLog(DEBUG, arg0, args...)
	}
}

// LogLogTrace logs a message at the trace log level.
func LogLogTrace(arg0 interface{}, args ...interface{}) {
	if loglog != nil {
		loglog.intLog(TRACE, arg0, args...)
	}
}

// LogLogInfo logs a message at the info log level.
func LogLogInfo(arg0 interface{}, args ...interface{}) {
	if loglog != nil {
		loglog.intLog(INFO, arg0, args...)
	}
}

// LogLogWarn logs a message at the warn log level.
func LogLogWarn(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0, args...)
	if loglog != nil {
		loglog.intLog(WARNING, msg)
	}
	return errors.New(msg)
}

// LogLogError logs a message at the error log level.
func LogLogError(arg0 interface{}, args ...interface{})  error {
	msg := FormatMessage(arg0, args...)
	if loglog != nil {
		loglog.intLog(ERROR, arg0, args...)
	}
	return errors.New(msg)
}
