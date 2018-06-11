// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
)

/****** Filters map ******/

type Filters map[string]*Filter

// Make a new filters
func NewFilters() *Filters {
	return &Filters{}
}

// Add a new filter to the filters map which will only log messages at lvl or
// higher. And call Appender.Init() to allow the appender protocol to perform 
// any initialization steps it needs.
// This function should be called before install filters to logger by Logger.SetFilters(fs)
// Returns the Filters for chaining.
func (fs Filters) Add(tag string, lvl Level, writer Appender) *Filters {
	if filt, ok := fs[tag]; ok {
		filt.Close()
		delete(fs, tag)
	}
	writer.Init()
	fs[tag] = NewFilter(lvl, writer)
	return &fs
}

// Get the filter with tag
func (fs Filters) Get(tag string) *Filter {
	if filt, ok := fs[tag]; ok {
		return filt
	}

	return nil
}

// Close and remove all filters in preparation for exiting the program or a
// reconfiguration of logging.  Calling this is not really imperative, unless
// you want to guarantee that all log messages are written.
// This function should be called after release filters by Logger.SetFilters(nil)
func (fs Filters) Close() {
	// Close all filters
	for tag, filt := range fs {
		filt.Close()
		delete(fs, tag)
	}
}

// Check log level
// Return skip or not
func (fs Filters) Skip (lvl Level) bool {
	for _, filt := range fs {
		if lvl >= filt.Level {
			return false
		}
	}
	return true
}

// Dispatch the logs
func (fs Filters) Dispatch(rec *LogRecord) {
	for _, filt := range fs {
		if rec.Level < filt.Level {
			continue
		}
		filt.writeToChan(rec)
	}
}
