package nxlog4go

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

	Data  map[string]interface{} // Contains all the fields set by the user.
	Index []string
}

// With adds key-value pairs to the log record.
func (r *Recorder) With(args ...interface{}) *Recorder {
	r.Data, r.Index, _ = ArgsToMap(args)
	return r
}

func (r *Recorder) WithMore(args ...interface{}) *Recorder {
	if len(args) == 0 {
		return r
	}

	data, index, _ := ArgsToMap(args)
	if len(data) <= 0 {
		return r
	}

	if r.Data == nil {
		r.Data = make(map[string]interface{}, len(args)/2)
	}
	for k, v := range data {
		r.Data[k] = v
	}
	r.Index = append(r.Index, index...)
	return r
}
