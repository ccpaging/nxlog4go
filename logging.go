// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
	"runtime"
	"time"

	"github.com/ccpaging/nxlog4go/driver"
)

// Log sends a log message with level and message.
// Call depth:
//  2 - Where calling the wrapper of logger.Log(...)
//  1 - Where calling logger.Log(...)
//  0 - Where internal calling entry.flush()
func (l *Logger) Log(calldepth int, level int, arg0 interface{}, args ...interface{}) {
	if !l.enabled(level) {
		return
	}

	r := &driver.Recorder{
		Prefix:  l.prefix,
		Level:   level,
		Message: driver.ArgsToString(arg0, args...),
		Created: time.Now(),
	}

	if l.caller {
		// Determine caller func - it's expensive.
		_, r.Source, r.Line, _ = runtime.Caller(calldepth)
	} else {
		r.Source, r.Line = "", 0
	}

	l.Dispatch(r)
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Finest(arg0 interface{}, args ...interface{}) {
	l.Log(2, FINEST, arg0, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Fine(arg0 interface{}, args ...interface{}) {
	l.Log(2, FINE, arg0, args...)
}

// Debug is a utility method for debug log messages.
// See ArgsToString for an explanation of the arguments.
func (l *Logger) Debug(arg0 interface{}, args ...interface{}) {
	l.Log(2, DEBUG, arg0, args...)
}

// Trace logs a message at the trace log level.
// See ArgsToString for an explanation of the arguments.
func (l *Logger) Trace(arg0 interface{}, args ...interface{}) {
	l.Log(2, TRACE, arg0, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (l *Logger) Info(arg0 interface{}, args ...interface{}) {
	l.Log(2, INFO, arg0, args...)
}

// Warn logs a message at the warn log level and returns the formatted error.
// At the warn level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See ArgsToString for further explanation of the arguments.
func (l *Logger) Warn(arg0 interface{}, args ...interface{}) error {
	msg := driver.ArgsToString(arg0, args...)
	l.Log(2, WARN, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and ArgsToString for an explanation
// of the parameters.
func (l *Logger) Error(arg0 interface{}, args ...interface{}) error {
	msg := driver.ArgsToString(arg0, args...)
	l.Log(2, ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and ArgsToString for an explanation
// of the parameters.
func (l *Logger) Critical(arg0 interface{}, args ...interface{}) error {
	msg := driver.ArgsToString(arg0, args...)
	l.Log(2, CRITICAL, msg)
	return errors.New(msg)
}
