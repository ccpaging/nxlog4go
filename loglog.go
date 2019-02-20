// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"os"
)

var loglog *Logger

// GetLogLog creates internal logger if not existed, and returns it.
// This logger used to output log statements from within the package.
// Do not set any filters.
func GetLogLog() *Logger {
	if loglog == nil {
		loglog = &Logger{
			out:    os.Stderr,
			level:  DEBUG,
			caller: true,
			prefix: "logg",
			layout: NewPatternLayout("%T %P %L %M\n"),
		}
	}
	return loglog
}

// LogLogDebug logs a message at the debug log level.
func LogLogDebug(args ...interface{}) {
	if loglog != nil {
		loglog.intLog(DEBUG, args...)
	}
}

// LogLogTrace logs a message at the trace log level.
func LogLogTrace(args ...interface{}) {
	if loglog != nil {
		loglog.intLog(TRACE, args...)
	}
}

// LogLogInfo logs a message at the info log level.
func LogLogInfo(args ...interface{}) {
	if loglog != nil {
		loglog.intLog(INFO, args...)
	}
}

// LogLogWarn logs a message at the warn log level.
func LogLogWarn(args ...interface{}) error {
	msg := FormatMessage(args...)
	if loglog != nil {
		loglog.intLog(WARNING, msg)
	}
	return errors.New(msg)
}

// LogLogError logs a message at the error log level.
func LogLogError(args ...interface{}) error {
	msg := FormatMessage(args...)
	if loglog != nil {
		loglog.intLog(ERROR, args...)
	}
	return errors.New(msg)
}
