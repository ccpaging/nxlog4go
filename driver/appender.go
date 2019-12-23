package driver

import (
	"errors"
)

// Appender is an interface for anything that should be able to write logs
type Appender interface {
	// Open opens and creates the appender.
	Open(dsn string, args ...interface{}) (Appender, error)

	// Set sets name-value option. Checkable
	Set(name string, value interface{}) error

	// Enabled returns true if the given recorder should be encoded
	// and written bytes after.
	//
	// Enabled can write the recorder directly with owner format,
	// and then return false.
	Enabled(*Recorder) bool

	// Write will be called to write the log bytes.
	Write([]byte) (int, error)

	// Close should clean up anything lingering about the Appender, as it is called before
	// the Appender is removed.  Write should not be called after Close.
	Close()
}

type NopAppender struct{}

func (*NopAppender) Open(string, ...interface{}) (Appender, error) { return &NopAppender{}, nil }
func (*NopAppender) Set(string, interface{}) error                 { return nil }
func (*NopAppender) Enabled(*Recorder) bool                        { return false }
func (*NopAppender) Write([]byte) (int, error)                     { return 0, errors.New("NOP") }
func (*NopAppender) Close()                                        {}

/** Register **/

var registered = make(map[string]Appender)

// Register is called by 3rd appender's init() function
// to register the new appender interface.
func Register(name string, app Appender) {
	if name == "" {
		return
	}
	if app == nil {
		delete(registered, name)
		return
	}
	registered[name] = app
}

// Open opens and creates the named appender.
//
// Return new appender and error
func Open(name string, dsn string, args ...interface{}) (Appender, error) {
	if app, ok := registered[name]; ok {
		return app.Open(dsn, args...)
	}
	return nil, errors.New("Not register " + name)
}
