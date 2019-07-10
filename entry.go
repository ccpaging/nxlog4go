// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// According to https://github.com/sirupsen/logrus/blob/master/entry.go

package nxlog4go

import (
	"errors"
	"runtime"
	"time"
)

/****** Entry ******/

// A Entry contains all of the pertinent information for each message
// It is the final or intermediate logging entry also. It contains all
// the fields passed with WithField{,s}. It's finally logged when Trace, Debug,
// Info, Warn, Error, Fatal or Panic is called on it.
type Entry struct {
	logger    *Logger
	calldepth int

	*LogRecord
}

// NewEntry creates a new logging entry with a logger
func NewEntry(l *Logger) *Entry {
	return &Entry{
		logger: l,
		LogRecord: &LogRecord{
			Prefix: l.prefix,
			Data:   make(map[string]interface{}, 6),
		},
	}
}

// With adds key-value pairs to the log record.
func (e *Entry) With(args ...interface{}) *Entry {
	if len(args) == 0 {
		return e
	}

	for i := 0; i < len(args); i += 2 {
		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			LogLogWarn("Ignored odd number argument. %d, key %v", i, args[i])
			break
		}
		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			LogLogWarn("Ignored key(not string). %d, key(%T) %v, value %v", i, key, key, val)
		} else {
			e.index = append(e.index, keyStr)
			switch v := val.(type) {
			case string:
				e.Data[keyStr] = val.(string)
			case error:
				e.Data[keyStr] = v.Error()
			case func() string:
				e.Data[keyStr] = v()
			default:
				e.Data[keyStr] = val
			}
		}
	}
	return e
}

func (e *Entry) flush() {
	l := e.logger
	l.mu.Lock()
	defer l.mu.Unlock()

	source, line := "", 0
	if l.caller {
		l.mu.Unlock()
		// Determine caller func - it's expensive.
		_, source, line, _ = runtime.Caller(e.calldepth)
		l.mu.Lock()
	}

	// Make the log record
	e.Created = time.Now()
	e.Source = source
	e.Line = line

	result := true
	if l.preHook != nil {
		result = l.preHook(l.out, e.LogRecord)
	}
	var (
		n   int
		err error
	)
	if result && l.out != nil && e.Level >= l.level {
		l.out.Write(l.layout.Format(e.LogRecord))
	}
	if l.postHook != nil {
		l.postHook(l.out, e.LogRecord, n, err)
	}

	if l.filters != nil {
		l.filters.Dispatch(e.LogRecord)
	}
}

// Log sends a log message with level, and message.
// Call depth:
//  2 - Where calling the wrapper of entry.Log(...)
//  1 - Where calling entry.Log(...)
//  0 - Where internal calling entry.flush()
func (e *Entry) Log(calldepth int, level Level, args ...interface{}) {
	if !e.logger.Skip(level) {
		e.Level = level
		if len(args) > 0 {
			e.Message = FormatMessage(args...)
		}
		e.calldepth = calldepth + 1
		e.flush()
	}
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Finest(args ...interface{}) {
	e.Log(2, FINEST, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Fine(args ...interface{}) {
	e.Log(2, FINE, args...)
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
//   When given anything else, the log message will be each of the arguments
//   formatted with %v and separated by spaces (ala Sprint).
func (e *Entry) Debug(args ...interface{}) {
	e.Log(2, DEBUG, args...)
}

// Trace logs a message at the trace log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Trace(args ...interface{}) {
	e.Log(2, TRACE, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (e *Entry) Info(args ...interface{}) {
	e.Log(2, INFO, args...)
}

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See Debug for further explanation of the arguments.
func (e *Entry) Warn(args ...interface{}) error {
	msg := FormatMessage(args...)
	e.Log(2, WARNING, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (e *Entry) Error(args ...interface{}) error {
	msg := FormatMessage(args...)
	e.Log(2, ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (e *Entry) Critical(args ...interface{}) error {
	msg := FormatMessage(args...)
	e.Log(2, CRITICAL, msg)
	return errors.New(msg)
}
