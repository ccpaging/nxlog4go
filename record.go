// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.
// According to https://github.com/sirupsen/logrus/blob/master/entry.go

package nxlog4go

import (
	"errors"
	"runtime"
	"time"
)

/****** LogRecord ******/

// A LogRecord contains all of the pertinent information for each message
// It is the final or intermediate logging entry also. It contains all
// the fields passed with WithField{,s}. It's finally logged when Trace, Debug,
// Info, Warn, Error, Fatal or Panic is called on it. These objects can be
// reused and passed around as much as you wish to avoid field duplication.
type LogRecord struct {
	logger *Logger

	Level   Level     // The log level
	Created time.Time // The time at which the log message was created (nanoseconds)
	Prefix  string    // The prefix
	Source  string    // The source
	Line    int       // The source line
	Message string    // The log message

	Data map[string]interface{} // Contains all the fields set by the user.
}

// NewLogRecord creates a new logger record with a logger
func NewLogRecord(l *Logger) *LogRecord {
	return &LogRecord{
		logger: l,
		Data:   make(map[string]interface{}, 6),
	}
}

// With adds key-value pairs to the log record.
func (r *LogRecord) With(args ...interface{}) *LogRecord {
	if len(args) == 0 {
		return r
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
			switch v := val.(type) {
			case string:
				r.Data[keyStr] = val.(string)
			case error:
				r.Data[keyStr] = v.Error()
			case func() string:
				r.Data[keyStr] = v()
			default:
				r.Data[keyStr] = val
			}
		}
	}
	return r
}

func (r *LogRecord) write(calldepth int) {
	l := r.logger
	l.mu.Lock()
	defer l.mu.Unlock()

	source, line := "", 0
	if l.caller {
		l.mu.Unlock()
		// Determine caller func - it's expensive.
		_, source, line, _ = runtime.Caller(calldepth)
		l.mu.Lock()
	}

	// Make the log record
	r.Created = time.Now()
	r.Prefix = l.prefix
	r.Source = source
	r.Line = line

	result := true
	if l.preHook != nil {
		result = l.preHook(l.out, r)
	}
	var (
		n   int
		err error
	)
	if result && l.out != nil && r.Level >= l.level {
		l.out.Write(l.layout.Format(r))
	}
	if l.postHook != nil {
		l.postHook(l.out, r, n, err)
	}

	if l.filters != nil {
		l.filters.Dispatch(r)
	}
}

// Log sends a log message with caller skip, level, and message.
func (r *LogRecord) intLog(level Level, args ...interface{}) {
	if !r.logger.Skip(level) {
		r.Level = level
		if len(args) > 0 {
			r.Message = FormatMessage(args...)
		}
		r.write(2 + 1)
	}
}

// Log sends a log message with level, and message.
func (r *LogRecord) Log(level Level, args ...interface{}) {
	r.intLog(level, args...)
}

// Finest logs a message at the finest log level.
// See Debug for an explanation of the arguments.
func (r *LogRecord) Finest(args ...interface{}) {
	r.intLog(FINEST, args...)
}

// Fine logs a message at the fine log level.
// See Debug for an explanation of the arguments.
func (r *LogRecord) Fine(args ...interface{}) {
	r.intLog(FINE, args...)
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
func (r *LogRecord) Debug(args ...interface{}) {
	r.intLog(DEBUG, args...)
}

// Trace logs a message at the trace log level.
// See Debug for an explanation of the arguments.
func (r *LogRecord) Trace(args ...interface{}) {
	r.intLog(TRACE, args...)
}

// Info logs a message at the info log level.
// See Debug for an explanation of the arguments.
func (r *LogRecord) Info(args ...interface{}) {
	r.intLog(INFO, args...)
}

// Warn logs a message at the warning log level and returns the formatted error.
// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closures are executed to format the error message.
// See Debug for further explanation of the arguments.
func (r *LogRecord) Warn(args ...interface{}) error {
	msg := FormatMessage(args...)
	r.intLog(WARNING, msg)
	return errors.New(msg)
}

// Error logs a message at the error log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (r *LogRecord) Error(args ...interface{}) error {
	msg := FormatMessage(args...)
	r.intLog(ERROR, msg)
	return errors.New(msg)
}

// Critical logs a message at the critical log level and returns the formatted error,
// See Warn for an explanation of the performance and Debug for an explanation
// of the parameters.
func (r *LogRecord) Critical(args ...interface{}) error {
	msg := FormatMessage(args...)
	r.intLog(CRITICAL, msg)
	return errors.New(msg)
}
