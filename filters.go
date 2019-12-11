// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

/****** Filters map ******/

// Filters represents a collection of Appenders through which log messages are
// written.
type Filters map[string]*Filter

// NewFilters creates a new filters
func NewFilters() Filters {
	return Filters{}
}

// Add a new filter to the filters map which will only log messages at level or
// higher. And call Appender.Init() to allow the appender protocol to perform
// any initialization steps it needs.
// This function should be called before install filters to logger by Logger.SetFilters(fs)
// Returns the Filters for chaining.
func (fs Filters) Add(tag string, level int, appender Appender) Filters {
	if filt, ok := fs[tag]; ok {
		filt.Close()
		delete(fs, tag)
	}
	fs[tag] = NewFilter(level, appender)
	return fs
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

// Skip checks log level and return whether skip or not
func (fs Filters) enabled(level int) bool {
	for _, filt := range fs {
		if level >= filt.Level {
			return true
		}
	}
	return false
}

// Dispatch the logs
func (fs Filters) Dispatch(r *Recorder) {
	for _, filt := range fs {
		if r.Level < filt.Level {
			continue
		}
		filt.writeToChan(r)
	}
}
