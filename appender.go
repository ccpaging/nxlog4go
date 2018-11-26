// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

/****** Appender ******/

// This is an interface for anything that should be able to write logs
type Appender interface {
	// Set option about the Layout. The options should be set as default.
	// Chainable.
	Set(k string, v interface{}) Appender

	// Set option about the Layout. The options should be set as default.
	// Checkable
	SetOption(k string, v interface{}) error

	// Init is called by the filters.Add() function to allow the appender
	// protocol to perform any initialization steps it needs. 
	// It should save the handle for future use, as well.
	Init()

	// This will be called to log a LogRecord message.
	Write(rec *LogRecord)

	// This should clean up anything lingering about the Appender, as it is called before
	// the Appender is removed.  Write should not be called after Close.
	Close()
}
