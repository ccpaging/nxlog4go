package driver

import (
	"fmt"
	"testing"
)

func TestRecorderWithError(t *testing.T) {
	r := &Recorder{}
	err := fmt.Errorf("kaboom at layer %d", 4711)
	if got := r.With("error", err).Fields["error"]; got != err.Error() {
		t.Errorf("With(\"%s\", \"%s\"):", "error", err)
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", err)
	}
}

func TestRecorderWithArgs(t *testing.T) {
	r := &Recorder{}
	err := fmt.Errorf("kaboom at layer %d", 4711)
	want := 3
	if got := len(r.With("error", err, "k1", "v1", "k2", "v2").Fields); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	} else if got = len(r.Index); got != want {
		t.Errorf("   got index %q", got)
		t.Errorf("  want index %q", want)
	}
}

func TestRecorderWithFunc(t *testing.T) {
	fn := func() string { return "Hello, world!" }

	r := &Recorder{}
	if got, ok := r.With("fn()", fn).Fields["fn()"]; !ok {
		t.Errorf("With(\"fn()\", fn) should not be ignored")
	} else if s, ok := got.(string); !ok {
		t.Errorf("Not string")
	} else if s != fn() {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", fn())
	}
}
