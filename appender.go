// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"errors"
)

/****** Appender ******/

// Appender is an interface for anything that should be able to write logs
type Appender interface {
	// Open opens and creates the appender.
	Open(dsn string, args ...interface{}) (Appender, error)

	// Set options about the appender. The options should be set as default.
	// Chainable.
	Set(args ...interface{}) Appender

	// SetOption about the appender. The options should be set as default.
	// Return error if option name or value is bad.
	SetOption(k string, v interface{}) error

	// Write will be called to log a Entry message.
	Write(e *Entry)

	// Close should clean up anything lingering about the Appender, as it is called before
	// the Appender is removed.  Write should not be called after Close.
	Close()
}

var registered = make(map[string]Appender)

// Register is called by 3rd appender's init() function
// to register appender interface.
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
// Return new appender and error
func Open(name string, dsn string, args ...interface{}) (Appender, error) {
	if app, ok := registered[name]; ok {
		return app.Open(dsn, args...)
	}
	return nil, errors.New("Not register " + name)
}

// Close closes all log filters in preparation for exiting the program.
// Calling this is not really imperative, unless you want to
// guarantee that all log messages are written.  Close() removes
// all filters (and thus all appenders) from the logger.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.filters != nil {
		l.filters.Close()
		l.filters = nil
	}
}

// Skip determines whether any logging will be skipped or not.
func (l *Logger) skip(level int) bool {
	if l.out != nil && level >= l.level {
		return false
	}

	if l.filters != nil {
		if l.filters.skip(level) == false {
			return false
		}
	}

	// l.out == nil and l.filters == nil
	// or level < l.Level
	return true
}
