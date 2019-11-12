// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

/****** Appender ******/

// Appender is an interface for anything that should be able to write logs
type Appender interface {
	// Set option about the appender. The options should be set as default.
	// Chainable.
	Set(k string, v interface{}) Appender

	// SetOption about the appender. The options should be set as default.
	// Return error if option name or value is bad.
	SetOption(k string, v interface{}) error

	// Write will be called to log a LogRecord message.
	Write(e *Entry)

	// Close should clean up anything lingering about the Appender, as it is called before
	// the Appender is removed.  Write should not be called after Close.
	Close()
}
