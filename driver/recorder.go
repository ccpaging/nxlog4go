package driver

import (
	"time"
)

// Recorder contains all of the pertinent information for each message
// It is the final or intermediate logging entry also. It contains all
// the fields passed with With(key, value, ...). It's finally logged
// when Trace, Debug, Info, Warn, Error, Fatal or Panic is called on it.
type Recorder struct {
	Prefix  string    // The message prefix
	Source  string    // The message source
	Line    int       // The source line
	Level   int       // The log level
	Message string    // The log message
	Created time.Time // The time at which the log message was created (nanoseconds)
	Values  []interface{}
}

// With sets sets values to the log record.
func (r *Recorder) With(args ...interface{}) *Recorder {
	r.Values = ArgsToValues(args...)
	return r
}

// WithMore appends more values to the log record.
func (r *Recorder) WithMore(args ...interface{}) *Recorder {
	values := ArgsToValues(args...)
	r.Values = append(r.Values, values...)
	return r
}

// Fields return the fields and index of the log record.
func (r *Recorder) Fields() (map[string]interface{}, []string) {
	return LazyArgsToMap(r.Values...)
}
