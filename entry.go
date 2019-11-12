// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// According to https://github.com/sirupsen/logrus/blob/master/entry.go

package nxlog4go

import (
	"errors"
	"runtime"
	"time"
	"fmt"
)

func ArgsToMap(args []interface{}) (map[string]interface{}, []string, error) {
	d := make(map[string]interface{}, 0)
	o := make([]string, 0)
	for i := 0; i < len(args); i += 2 {
		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			return d, o, fmt.Errorf("The arguments' count (%d) should be odd. %v", len(args), args)
		}
		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if s, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			return d, o, fmt.Errorf("No.%d key should be string. %v", i, args)
		} else {
			o = append(o, s)
			switch v := val.(type) {
			case string:
				d[s] = val.(string)
			case error:
				d[s] = v.Error()
			case func() string:
				d[s] = v()
			default:
				d[s] = val
			}
		}
	}
	return d, o, nil
}

/****** Entry ******/

// A Entry contains all of the pertinent information for each message
// It is the final or intermediate logging entry also. It contains all
// the fields passed with WithField{,s}. It's finally logged when Trace, Debug,
// Info, Warn, Error, Fatal or Panic is called on it.
type Entry struct {
	Level   Level       // The log level
	Created time.Time // The time at which the log message was created (nanoseconds)
	Prefix  string    // The message prefix
	Source  string    // The message source
	Line    int       // The source line
	Message string    // The log message

	Data  map[string]interface{} // Contains all the fields set by the user.
	index []string

	logger    *Logger
	calldepth int
}

// NewEntry creates a new logging entry with a logger
func NewEntry(l *Logger) *Entry {
	return &Entry{
		Prefix: l.prefix,
		logger: l,
	}
}

// With adds key-value pairs to the log record.
func (e *Entry) With(args ...interface{}) *Entry {
	if len(args) == 0 {
		return e
	}

	if e.Data == nil {
		e.Data = make(map[string]interface{}, 0)
	}
	data, index, err := ArgsToMap(args)
	if err != nil {
		LogLogWarn(err)
	}
	for k, v := range data {
		e.Data[k] = v
	}
	e.index = append(e.index, index ...)
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
		result = l.preHook(e)
	}
	var (
		n   int
		err error
	)
	if result && l.out != nil && e.Level >= l.level {
		l.out.Write(l.layout.Format(e))
	}
	if l.postHook != nil {
		l.postHook(e, n, err)
	}

	if l.filters != nil {
		l.filters.Dispatch(e)
	}
}

// Log sends a log message with level, and message.
// Call depth:
//  2 - Where calling the wrapper of entry.Log(...)
//  1 - Where calling entry.Log(...)
//  0 - Where internal calling entry.flush()
func (e *Entry) Log(calldepth int, level Level, arg0 interface{}, args ...interface{}) {
	if e.logger.Skip(level) {
		return
	}

	e.Level = level
	e.Message = FormatMessage(arg0)
	e.With(args...)
	e.calldepth = calldepth + 1
	e.flush()
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

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See Debug for further explanation of the arguments.
func (e *Entry) Warn(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0)
	e.Log(2, WARNING, msg, args ...)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (e *Entry) Error(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0)
	e.Log(2, ERROR, msg, args...)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (e *Entry) Critical(arg0 interface{}, args ...interface{}) error {
	msg := FormatMessage(arg0)
	e.Log(2, CRITICAL, msg, args...)
	return errors.New(msg)
}
