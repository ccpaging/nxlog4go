// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
    "fmt"
	"os"
	"time"
)

type FilterConfig struct {
	Enabled  string        `xml:"enabled,attr"`
	Tag      string        `xml:"tag"`
	Level    string        `xml:"level"`
	Type     string        `xml:"type"`
	Properties []AppenderProp `xml:"property"`
}

type LoggerConfig struct {
	FilterConfigs []FilterConfig `xml:"filter"`
}

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
		fmt.Fprintf(os.Stderr, "Filter: channel has been closed. Message is [%s]\n", rec.Message)
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

// Check filter's configuration
func CheckFilterConfig(fc FilterConfig) (ok bool, enabled bool, lvl Level) {
	ok, enabled, lvl = true, false, INFO

	// Check required children
	if len(fc.Enabled) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required attribute %s\n", "enabled")
		ok = false
	} else {
		enabled = fc.Enabled != "false"
	}
	if len(fc.Tag) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required child <%s>\n", "tag")
		ok = false
	}
	if len(fc.Type) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required child <%s>\n", "type")
		ok = false
	}
	if len(fc.Level) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required child <%s>\n", "level")
		ok = false
	}

	switch fc.Level {
	case "FINEST":
		lvl = FINEST
	case "FINE":
		lvl = FINE
	case "DEBUG":
		lvl = DEBUG
	case "TRACE":
		lvl = TRACE
	case "INFO":
		lvl = INFO
	case "WARNING":
		lvl = WARNING
	case "ERROR":
		lvl = ERROR
	case "CRITICAL":
		lvl = CRITICAL
	default:
		fmt.Fprintf(os.Stderr, 
			"CheckFilterConfig: Required child level for filter has unknown value. %s\n", 
			fc.Level)
		ok = false
	}
	return ok, enabled, lvl
}


/****** Filters map ******/

type Filters map[string]*Filter

// Make a new filters
func NewFilters() *Filters {
	return &Filters{}
}

// Add a new filter to the filters map which will only log messages at lvl or
// higher.  This function should not be called from multiple goroutines.
// Returns the logger for chaining.
func (fs Filters) Add(name string, lvl Level, writer Appender) *Filters {
	if filt, isExist := fs[name]; isExist {
		filt.Close()
		delete(fs, name)
	}
	fs[name] = NewFilter(lvl, writer)
	return &fs
}

// Close and remove all filters in preparation for exiting the program or a
// reconfiguration of logging.  Calling this is not really imperative, unless
// you want to guarantee that all log messages are written.  Close removes
// all filters (and thus all Appenders) from the logger.
// Returns the logger for chaining.
func (fs Filters) Close() {
	// Close all filters
	for name, filt := range fs {
		filt.Close()
		delete(fs, name)
	}
}
