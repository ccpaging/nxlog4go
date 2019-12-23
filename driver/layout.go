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

type NopLayout struct{}

func (*NopLayout) Set(string, interface{}) error { return nil }
func (*NopLayout) Encode(string, interface{})    {}
