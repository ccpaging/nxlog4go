// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"

	"github.com/ccpaging/nxlog4go/driver"
)

var loglog *Logger

// FormatLogLog is format for internal log
var FormatLogLog = "%T %P %L %M"

// GetLogLog creates internal logger if not existed, and returns it.
// This logger used to output log statements from within the package.
// Do not set any filters.
func GetLogLog() *Logger {
	if loglog == nil {
		loglog = NewLogger(DEBUG).SetOptions(
			"prefix", "logg",
			"format", FormatLogLog)
	}
	return loglog
}

// LogLogDebug logs a message at the debug log level.
func LogLogDebug(arg0 interface{}, args ...interface{}) {
	if loglog != nil {
		loglog.Log(2, DEBUG, arg0, args...)
	}
}

// LogLogTrace logs a message at the trace log level.
func LogLogTrace(arg0 interface{}, args ...interface{}) {
	if loglog != nil {
		loglog.Log(2, TRACE, arg0, args...)
	}
}

// LogLogInfo logs a message at the info log level.
func LogLogInfo(arg0 interface{}, args ...interface{}) {
	if loglog != nil {
		loglog.Log(2, INFO, arg0, args...)
	}
}

// LogLogWarn logs a message at the warn log level.
func LogLogWarn(arg0 interface{}, args ...interface{}) error {
	msg := driver.ArgsToString(arg0, args...)
	if loglog != nil {
		loglog.Log(2, WARN, msg)
	}
	return errors.New(msg)
}

// LogLogError logs a message at the error log level.
func LogLogError(arg0 interface{}, args ...interface{}) error {
	msg := driver.ArgsToString(arg0, args...)
	if loglog != nil {
		loglog.Log(2, ERROR, msg)
	}
	return errors.New(msg)
}
