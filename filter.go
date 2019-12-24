// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"

	"github.com/ccpaging/nxlog4go/driver"
)

// Filter contains:
//  - Enabler, the Enabler interface for filter log Recorder
//  - Layout, the Layout interface for encoding log Recorder
//  - Apps, the slice of the Appender interface
type Filter struct {
	driver.Enabler
	driver.Layout
	Apps []driver.Appender
}

// NewFilter creates a new filter with an enabler, a layout
// and appenders.
func NewFilter(enb driver.Enabler, lo driver.Layout, apps ...driver.Appender) *Filter {
	return &Filter{enb, lo, apps}
}

// Dispatch filters, encodes a log recorder to bytes, and writes it to all appenders.
//  - Enabler.Enabled, filter log Recorder.
//  - Layout.Encode, encode log Recorder to bytes.Buffer.
//  - Apps[i].Enabled, filter log recorder by appender.
//  - Apps[i].Write, append with log recorder encoded bytes.
func (f *Filter) Dispatch(r *driver.Recorder) {
	if f.Enabler != nil && !f.Enabler.Enabled(r) {
		return
	}

	out := new(bytes.Buffer)
	encoded := false
	for _, a := range f.Apps {
		if a != nil && !a.Enabled(r) {
			continue
		}
		if f.Layout == nil {
			continue
		}
		if !encoded {
			f.Layout.Encode(out, r)
			encoded = true
		}
		a.Write(out.Bytes())
	}
}

// Close closes all log appenders in preparation for exiting the program.
// Calling this is not really imperative, unless you want to
// guarantee that all log messages are written.  Close() removes
// all appenders from the filter.
func (f *Filter) Close() {
	for _, a := range f.Apps {
		if a != nil {
			a.Close()
		}
	}
	f.Apps = nil
}
