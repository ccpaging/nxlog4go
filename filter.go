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
// 
// DEPRECATED: Use appender owned level instead.
type Filter struct {
	entry    chan *Entry // write queue
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	Level int
	Appender
}

// NewFilter creates a new filter.
func NewFilter(level int, writer Appender) *Filter {
	f := &Filter{
		entry: make(chan *Entry, LogBufferLength),

		Level:    level,
		Appender: writer,
	}

	return f
}

// This is the filter's output method. This will block if the output
// buffer is full.
func (f *Filter) writeToChan(e *Entry) {
	f.runOnce.Do(func() {
		f.waitExit = &sync.WaitGroup{}
		f.waitExit.Add(1)
		go f.run(f.waitExit)
	})

	// Write after closed
	if f.waitExit == nil {
		f.Write(e)
		return
	}

	f.entry <- e
}

func (f *Filter) run(waitExit *sync.WaitGroup) {
	for {
		select {
		case e, ok := <-f.entry:
			if !ok {
				// drain channel
				for left := range f.entry {
					f.Write(left)
				}
				waitExit.Done()
				return
			}
			f.Write(e)
		}
	}
}

// Close the filter
func (f *Filter) Close() {
	if f.waitExit == nil {
		return
	}

	defer f.Appender.Close()

	// notify closing. See run()
	close(f.entry)
	// waiting for running channel closed
	f.waitExit.Wait()
	f.waitExit = nil
	// drain channel
	for e := range f.entry {
		f.Write(e)
	}
}
