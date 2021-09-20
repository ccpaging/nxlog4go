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

	Fields map[string]interface{} // Contains all the fields set by the user.
	Index  []string

	Values []interface{}
}

// WithValues sets values to the log record.
func (r *Recorder) WithValues(vals ...interface{}) *Recorder {
	r.Values = ArgsToValues(vals...)
	return r
}

// WithMoreValues appends more values to the log record.
func (r *Recorder) WithMoreValues(vals ...interface{}) *Recorder {
	values := ArgsToValues(vals...)
	r.Values = append(r.Values, values...)
	return r
}

// With sets name-value pairs to the log record.
func (r *Recorder) With(args ...interface{}) *Recorder {
	r.Fields, r.Index, _ = ArgsToMap(args...)
	return r
}

// WithMore appends name-value pairs to the log record.
func (r *Recorder) WithMore(args ...interface{}) *Recorder {
	if r.Fields == nil {
		r.Fields = make(map[string]interface{}, len(args)/2)
	}

	if len(args) == 0 {
		return r
	}

	fields, index, _ := ArgsToMap(args...)
	if len(fields) <= 0 {
		return r
	}
	for k, v := range fields {
		r.Fields[k] = v
	}
	r.Index = append(r.Index, index...)
	return r
}
