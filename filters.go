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
// higher. And call Appender.Init() to allow the appender protocol to perform 
// any initialization steps it needs.
// This function should be called before install filters to logger by Logger.SetFilters(fs)
// Returns the Filters for chaining.
func (fs Filters) Add(tag string, lvl Level, writer Appender) *Filters {
	if filt, isExist := fs[tag]; isExist {
		filt.Close()
		delete(fs, tag)
	}
	writer.Init()
	fs[tag] = NewFilter(lvl, writer)
	return &fs
}

func (fs Filters) Preload(tag string, writer Appender) *Filters {
	return fs.Add(tag, SILENT, writer)
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
		if len(fc.Tag) == 0 {
			LogLogError("config", "Required child tag")
			continue
		}
		tag := fc.Tag
		if len(fc.Level) == 0 {
			LogLogError("config", "Required child level")
			continue
		}
		lvl := GetLevel(fc.Level)
		if lvl >= SILENT {
			continue
		}
		filt, isExist := fs[tag]
		if !isExist {
			LogLogError("config", "Appender <%s> is not pre-installed", tag)
			continue
		}
		if filt.Appender == nil {
			LogLogError("config", "Appender <%s> pre-installed is nil", tag)
			continue
		}
		if !AppenderConfigure(filt.Appender, fc.Props) {
			continue
		}
		filt.Level = lvl
	}
	// Close and delete the appenders at SILENT
	for tag, filt := range fs {
		if filt.Level >= SILENT {
			filt.Close()
			delete(fs, tag)
		}
	}
}
