// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"github.com/ccpaging/nxlog4go/driver"
)

// AddFilter adds the named appenders to the Logger which will only log messages at
// or above level. This function should not be called from multiple goroutines.
//
// Returns the logger for chaining.
func (l *Logger) AddFilter(name string, level int, apps ...driver.Appender) *Logger {
	f := &driver.Filter{
		Name:    name,
		Enabler: driver.AtAbove(level),
		Layout:  nil,
		Apps:    apps,
	}
	return l.Attach(f)
}

// Attach adds the filters to logger.
func (l *Logger) Attach(filters ...*driver.Filter) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range filters {
		if f == nil {
			continue
		}
		if _, ok := l.filters[f.Name]; ok {
			// Existed
			continue
		}
		l.filters[f.Name] = f
	}

	return l
}

// Detach removes the filters from logger.
func (l *Logger) Detach(filters ...*driver.Filter) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, f := range filters {
		if f == nil {
			continue
		}
		if _, ok := l.filters[f.Name]; ok {
			// Existed
			delete(l.filters, f.Name)
		}
	}

	return l
}
