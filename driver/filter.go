// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package driver

import (
	"bytes"
)

// Filter contains:
//  - Enabler, the Enabler interface for filter log Recorder
//  - Layout, the Layout interface for encoding log Recorder
//  - Apps, the slice of the Appender interface
type Filter struct {
	Name string
	Enabler
	Layout
	Apps []Appender
}

// Dispatch filters, encodes a log recorder to bytes, and writes it to all appenders.
//  - Enabler.Enabled, filter log Recorder.
//  - Layout.Encode, encode log Recorder to bytes.Buffer.
//  - Apps[i].Enabled, filter log recorder by appender.
//  - Apps[i].Write, append with log recorder encoded bytes.
func (f *Filter) Dispatch(r *Recorder) {
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
// guarantee that all log messages are written.
//
// Notice: Close() removes all appenders from the filter.
func (f *Filter) Close() {
	for _, a := range f.Apps {
		if a != nil {
			a.Close()
		}
	}
	f.Apps = nil
}
