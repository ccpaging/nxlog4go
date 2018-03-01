// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
)

/****** Filters map ******/

type Filters map[string]*Filter

type FilterConfig struct {
	Tag		 string     	`xml:"tag"`
	Level    string     	`xml:"level"`
	Props	 []AppenderProp `xml:"property" json:"properties"`
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

func (fs Filters) Preload(tag string, writer Appender) *Filters {
	return fs.Add(tag, _SILENT_, writer)
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

func (fs Filters) LoadConfiguration(fcs []FilterConfig) {
	for _, fc := range fcs {
		ok, tag, lvl := getFilterConfig(fc)
		if !ok {
			continue
		}
		if lvl >= _SILENT_ {
			continue
		}
		filt, isExist := fs[tag]
		if !isExist {
			loglogError("LoadConfiguration", "Appender <%s> is not pre-installed", tag)
			continue
		}
		if filt.Appender == nil {
			loglogError("LoadConfiguration", "Appender <%s> pre-installed is nil", tag)
			continue
		}
		if !AppenderConfigure(filt.Appender, fc.Props) {
			continue
		}
		filt.Level = lvl
	}
	// Close and delete the appenders at _SILENT_
	for tag, filt := range fs {
		if filt.Level >= _SILENT_ {
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
		loglogError("getFilterConfig", "Required child <%s>", "tag")
		ok = false
	}
	if len(fc.Level) == 0 {
		loglogError("getFilterConfig", "Required child <%s>", "level")
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
	case "DISABLE", "DISA", "SILENT", "QUIET":
		lvl = _SILENT_
	default:
		lvl = _SILENT_
	}
	return ok, fc.Tag, lvl
}
