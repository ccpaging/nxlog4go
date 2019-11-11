// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"sync"
)

/****** Filter ******/

// Filter represents the log level below which no log records are written to
// the associated Appender.
type Filter struct {
	Level Level
	Appender

    runOnce sync.Once
	running *chan struct{} // Notify exited looping
	rec     chan *LogRecord // write queue
}

// NewFilter creates a new filter.
func NewFilter(level Level, writer Appender) *Filter {
	f := &Filter{
		Level:    level,
		Appender: writer,

		rec:     make(chan *LogRecord, LogBufferLength),
	}

	return f
}

// This is the filter's output method. This will block if the output
// buffer is full.
func (f *Filter) writeToChan(rec *LogRecord) {
    f.runOnce.Do(func() {
		ready := make(chan struct{})
		running := make(chan struct{})
		f.running = &running
		go f.run(ready, running)
		<-ready
	})
	
    if f.running == nil {
		f.Write(rec)
		return
	}
    
	f.rec <- rec
}

func (f *Filter) run(ready chan struct{}, running chan struct{}) {
	close(ready)
	for {
		select {
		case rec, ok := <-f.rec:
			if !ok {
                // drain channel
				for left := range f.rec {
					f.Write(left)
				}
				close(running)
				return
			}
			f.Write(rec)
		}
	}
}

// Close the filter
func (f *Filter) Close() {
	if f.running == nil {
		return
	}

	defer f.Appender.Close()

	// Notify log appender closing
	close(f.rec)
    // Waiting for running channel closed
	<-(*f.running)
	f.running = nil
}
