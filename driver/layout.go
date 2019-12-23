package driver

import (
	"bytes"
)

// Layout is is an interface for formatting log record
type Layout interface {
	// Set sets name-value option. Checkable.
	Set(name string, v interface{}) error

	// Encode encodes a log Recorder to bytes.
	Encode(out *bytes.Buffer, r *Recorder) int
}

type nopLayout struct{}

// NewNopLayoutwN returns a no-op Layout.
func NewNopLayout() Layout                             { return &nopLayout{} }
func (*nopLayout) Set(string, interface{}) error       { return nil }
func (*nopLayout) Encode(*bytes.Buffer, *Recorder) int { return 0 }
