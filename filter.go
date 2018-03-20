// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"time"
)

/****** Filter ******/

// A Filter represents the log level below which no log records are written to
// the associated Appender.
type Filter struct {
	Level Level
	Appender

	rec 	chan *LogRecord	// write queue
	closing	bool	// true if filter was closed at API level
}

// Create a new filter
func NewFilter(lvl Level, writer Appender) *Filter {
	f := &Filter {
		Level:		lvl,
		Appender:	writer,

		rec: 		make(chan *LogRecord, LogBufferLength),
		closing: 	false,
	}

	go f.run()
	return f
}

// This is the filter's output method. This will block if the output
// buffer is full. 
func (f *Filter) writeToChan(rec *LogRecord) {
	if f.closing {
		LogLogError("Filter", "Channel has been closed. Message is [%s]", rec.Message)
		return
	}
	f.rec <- rec
}

func (f *Filter) run() {
	for {
		select {
		case rec, ok := <-f.rec:
			if !ok {
				return
			}
			f.Write(rec)
		}
	}
}

// Close the filter
func (f *Filter) Close() {
	if f.closing {
		return
	}
	// sleep at most one second and let go routine running
	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closing channel
		time.Sleep(100 * time.Millisecond)
		if len(f.rec) <= 0 {
			break
		}
	}

	// block write channel
	f.closing = true

	defer f.Appender.Close()

	// Notify log appender closing
	close(f.rec)

	if len(f.rec) <= 0 {
		return
	}
	// drain the log channel and write direct
	for rec := range f.rec {
		f.Write(rec)
	}
}
