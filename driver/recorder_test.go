package driver

import (
	"fmt"
	"testing"
)

func TestRecorderWithFields(t *testing.T) {
	r := &Recorder{}
	want := 0
	if got := len(r.With().Fields); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	} else if got = len(r.Index); got != want {
		t.Errorf("   got index %q", got)
		t.Errorf("  want index %q", want)
	}
	if got := len(r.WithMore().Fields); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	} else if got = len(r.Index); got != want {
		t.Errorf("   got index %q", got)
		t.Errorf("  want index %q", want)
	}

	err := fmt.Errorf("kaboom at layer %d", 4711)
	want = 3
	if got := len(r.With("error", err, "k1", "v1", "k2", "v2").Fields); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	} else if got = len(r.Index); got != want {
		t.Errorf("   got index %q", got)
		t.Errorf("  want index %q", want)
	}
}

func TestRecorderWithFieldsError(t *testing.T) {
	r := &Recorder{}
	err := fmt.Errorf("kaboom at layer %d", 4711)
	if got := r.With("error", err).Fields["error"]; got != err.Error() {
		t.Errorf("With(\"%s\", \"%s\"):", "error", err)
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", err)
	}
}

func TestRecorderWithFieldsFunc(t *testing.T) {
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

func TestRecorderWithValues(t *testing.T) {
	r := &Recorder{}
	want := 0
	if got := len(r.WithValues().Values); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	if got := len(r.WithMoreValues().Values); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	err := fmt.Errorf("kaboom at layer %d", 4711)
	want = 5
	if got := len(r.WithValues("error", err, "v1", "v2", "v3").Values); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestRecorderWithValuesFunc(t *testing.T) {
	fn := func() string { return "Hello, world!" }

	r := &Recorder{}
	values := r.WithValues(fn).Values
	if len(values) <= 0 {
		t.Errorf("With(\"fn()\", fn) should not be ignored")
	} else if got, ok := values[0].(string); !ok {
		t.Errorf("Not string")
	} else if got != fn() {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", fn())
	}
}

func TestRecorderWithValuesFunc2(t *testing.T) {
	want := "Hello, world!"
	fn := func() (int, string) { return 1, want }

	r := &Recorder{}
	values := r.WithValues(fn()).Values
	if len(values) <= 1 {
		t.Errorf("With(\"fn()\", fn) should not be ignored")
	} else if got, ok := values[1].(string); !ok {
		t.Errorf("Not string")
	} else if got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}
