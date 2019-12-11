// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// According to https://github.com/sirupsen/logrus/blob/master/entry.go

package nxlog4go

import (
	"bytes"
	"errors"
	"runtime"
	"time"
)

// A Entry contains all of the pertinent information for each message
// It is the final or intermediate logging entry also. It contains all
// the fields passed with WithField{,s}. It's finally logged when Trace, Debug,
// Info, Warn, Error, Fatal or Panic is called on it.
type Entry struct {
	r *Recorder
	l *Logger
}

// NewEntry creates a new logging entry with a logger
func NewEntry(l *Logger) *Entry {
	return &Entry{
		r: &Recorder{
			Prefix: l.prefix,
		},
		l: l,
	}
}

// With adds key-value pairs to the log record.
func (e *Entry) With(args ...interface{}) *Entry {
	e.r.With(args...)
	return e
}

func (e *Entry) WithMore(args ...interface{}) *Entry {
	e.r.WithMore(args...)
	return e
}

// Log sends a log message with level, and message.
// Call depth:
//  2 - Where calling the wrapper of entry.Log(...)
//  1 - Where calling entry.Log(...)
//  0 - Where internal calling entry.flush()
func (e *Entry) Log(calldepth int, level int, arg0 interface{}, args ...interface{}) {
	r, l := e.r, e.l
	if !l.enabled(level) {
		return
	}
	r.Level = level
	r.Message = ArgsToString(arg0)
	r.WithMore(args...)
	r.Created = time.Now()

	if l.caller {
		// Determine caller func - it's expensive.
		_, r.Source, r.Line, _ = runtime.Caller(calldepth)
	} else {
		r.Source, r.Line = "", 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.out != nil && level >= l.level {
		buf := new(bytes.Buffer)
		l.layout.Encode(buf, r)
		l.out.Write(buf.Bytes())
	}

	// Dispatch to all appender
	if l.filters != nil {
		l.filters.Dispatch(r)
	}
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Finest(arg0 interface{}, args ...interface{}) {
	e.Log(2, FINEST, arg0, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Fine(arg0 interface{}, args ...interface{}) {
	e.Log(2, FINE, arg0, args...)
}

// Debug is a utility method for debug log messages.
// The behavior of Debug depends on the first argument:
// - arg0 is a string
//   When given a string as the first argument, this behaves like Logf but with
//   the DEBUG log level: the first argument is interpreted as a format for the
//   latter arguments.
// - arg0 is a func()string
//   When given a closure of type func()string, this logs the string returned by
//   the closure iff it will be logged.  The closure runs at most one time.
// - arg0 is interface{}
//   When given anything else, formatted as string (ala Sprint).
func (e *Entry) Debug(arg0 interface{}, args ...interface{}) {
	e.Log(2, DEBUG, arg0, args...)
}

// Trace logs a message at the trace log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Trace(arg0 interface{}, args ...interface{}) {
	e.Log(2, TRACE, arg0, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Info(arg0 interface{}, args ...interface{}) {
	e.Log(2, INFO, arg0, args...)
}

// Warn logs a message at the warn log level and returns the formatted error.
// At the warn level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See Debug for further explanation of the arguments.
func (e *Entry) Warn(arg0 interface{}, args ...interface{}) error {
	msg := ArgsToString(arg0)
	e.Log(2, WARN, msg, args...)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (e *Entry) Error(arg0 interface{}, args ...interface{}) error {
	msg := ArgsToString(arg0)
	e.Log(2, ERROR, msg, args...)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (e *Entry) Critical(arg0 interface{}, args ...interface{}) error {
	msg := ArgsToString(arg0)
	e.Log(2, CRITICAL, msg, args...)
	return errors.New(msg)
}
