// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
    "fmt"
	"os"
)

/****** Filters map ******/

type Filters map[string]*Filter

type FilterConfig struct {
	Tag		 string     	`xml:"tag"`
	Level    string     	`xml:"level"`
	Props	 []AppenderProp `xml:"property"`
}

type LoggerConfig struct {
	Filters  []FilterConfig `xml:"filter" json:"filters"`
}

// Make a new filters
func NewFilters() *Filters {
	return &Filters{}
}

// Add a new filter to the filters map which will only log messages at lvl or
// higher.
// Returns the Filters for chaining.
// This function should be called before install filters to logger by Logger.SetFilters(fs)
func (fs Filters) Add(tag string, lvl Level, writer Appender) *Filters {
	if filt, isExist := fs[tag]; isExist {
		filt.Close()
		delete(fs, tag)
	}
	fs[tag] = NewFilter(lvl, writer)
	return &fs
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

func (fs Filters) LoadConfiguration(fcs []FilterConfig) {
	for _, fc := range fcs {
		ok, tag, lvl := getFilterConfig(fc)
		if !ok {
			continue
		}
		if lvl >= OFFLevel {
			continue
		}
		filt, isExist := fs[tag]
		if !isExist {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Appender <%s> is not pre-installed\n", tag)
			continue
		}
		if filt.Appender == nil {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Appender <%s> pre-installed is nil\n", tag)
			continue
		}
		if !AppenderConfigure(filt.Appender, fc.Props) {
			continue
		}
		filt.Level = lvl
	}
	// Close and delete the appenders at OFFLevel
	for tag, filt := range fs {
		if filt.Level >= OFFLevel {
			filt.Close()
			delete(fs, tag)
		}
	}
}

// Check filter's configuration
func getFilterConfig(fc FilterConfig) (ok bool, tag string, lvl Level) {
	ok, tag, lvl = true, "", INFO

	// Check required children
	if len(fc.Tag) == 0 {
		fmt.Fprintf(os.Stderr, "getFilterConfig: Required child <%s>\n", "tag")
		ok = false
	}
	if len(fc.Level) == 0 {
		fmt.Fprintf(os.Stderr, "getFilterConfig: Required child <%s>\n", "level")
		ok = false
	}

	switch fc.Level {
	case "FINEST", "FNST":
		lvl = FINEST
	case "FINE":
		lvl = FINE
	case "DEBUG", "DEBG":
		lvl = DEBUG
	case "TRACE", "TRAC":
		lvl = TRACE
	case "INFO":
		lvl = INFO
	case "WARNING", "WARN":
		lvl = WARNING
	case "ERROR", "EROR":
		lvl = ERROR
	case "CRITICAL", "CRIT":
		lvl = CRITICAL
	case "OFF", "OFFL", "DISABLE", "DISA":
		lvl = OFFLevel
	default:
		lvl = OFFLevel
	}
	return ok, fc.Tag, lvl
}
