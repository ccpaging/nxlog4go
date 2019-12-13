// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
)

// A Filter contains an int level, a Layout and multi appenders.
type Filter struct {
	level   int
	encoder Layout
	apps    []Appender
}

// NewFilter creates a new filter with an int level, a layout
// and appenders.
func NewFilter(level int, enco Layout, apps ...Appender) *Filter {
	return &Filter{level: level, encoder: enco, apps: apps}
}

// Close closes all log appenders in preparation for exiting the program.
// Calling this is not really imperative, unless you want to
// guarantee that all log messages are written.  Close() removes
// all appenders from the filter.
func (f *Filter) Close() {
	for _, a := range f.apps {
		if a != nil {
			a.Close()
		}
	}
	f.apps = nil
}

// Dispatch encodes a log recorder to bytes and writes it to all appenders.
func (f *Filter) Dispatch(r *Recorder) {
	if r.Level < f.level {
		return
	}

	out := new(bytes.Buffer)
	enco := false
	for _, a := range f.apps {
		if !a.Enabled(r) {
			continue
		}
		if f.encoder == nil {
			continue
		}
		if !enco {
			f.encoder.Encode(out, r)
			enco = true
		}
		a.Write(out.Bytes())
	}
}

func closeFilters(filters []*Filter) {
	for _, f := range filters {
		if f != nil {
			f.Close()
		}
	}
}
func findFilter(filters []*Filter, filter *Filter) int {
	for i, f := range filters {
		if f == filter {
			return i
		}
	}
	return -1
}

// Attach adds the filters to logger.
func (l *Logger) Attach(filters ...*Filter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range filters {
		if f == nil {
			continue
		}
		if i := findFilter(l.filters, f); i >= 0 {
			// Existed
			continue
		}
		l.filters = append(l.filters, f)
	}
}

// Detach removes the filters from logger.
func (l *Logger) Detach(filters ...*Filter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range filters {
		if f == nil {
			continue
		}
		if i := findFilter(l.filters, f); i >= 0 {
			// Existed
			l.filters = append(l.filters[:i], l.filters[i+1:]...)
		}
	}
}

// Dispatch encodes a log recorder to bytes and writes it.
func (l *Logger) Dispatch(r *Recorder) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.out != nil && r.Level >= l.level {
		buf := new(bytes.Buffer)
		l.layout.Encode(buf, r)
		l.out.Write(buf.Bytes())
	}

	for _, f := range l.filters {
		if f != nil {
			f.Dispatch(r)
		}
	}
}
