// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"sync"
)

/****** Variables ******/

var (
	// LogBufferLength specifies how many log messages a particular log4go
	// logger can buffer at a time before writing them.
	LogBufferLength = 32
)

/****** Filter ******/

// Filter represents the log level below which no log entry are written to
// the associated Appender.
type Filter struct {
	Level int
	Appender

	runOnce sync.Once
	running *chan struct{} // Notify exited looping
	entry   chan *Entry    // write queue
}

// NewFilter creates a new filter.
func NewFilter(level int, writer Appender) *Filter {
	f := &Filter{
		Level:    level,
		Appender: writer,

		entry: make(chan *Entry, LogBufferLength),
	}

	return f
}

// This is the filter's output method. This will block if the output
// buffer is full.
func (f *Filter) writeToChan(e *Entry) {
	f.runOnce.Do(func() {
		ready := make(chan struct{})
		running := make(chan struct{})
		f.running = &running
		go f.run(ready, running)
		<-ready
	})

	if f.running == nil {
		f.Write(e)
		return
	}

	f.entry <- e
}

func (f *Filter) run(ready chan struct{}, running chan struct{}) {
	close(ready)
	for {
		select {
		case e, ok := <-f.entry:
			if !ok {
				// drain channel
				for left := range f.entry {
					f.Write(left)
				}
				close(running)
				return
			}
			f.Write(e)
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
	close(f.entry)
	// Waiting for running channel closed
	<-(*f.running)
	f.running = nil
}
