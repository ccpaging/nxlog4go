package driver

import (
	"fmt"
	"testing"
)

func TestRecorderWithFields(t *testing.T) {
	r := &Recorder{}
	fields, index := r.Fields()
	if got, want := len(fields), 0; got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	} else if got = len(index); got != want {
		t.Errorf("   got index %q", got)
		t.Errorf("  want index %q", want)
	}

	fields, index = r.WithMore().Fields()
	if got, want := len(fields), 0; got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	} else if got = len(index); got != want {
		t.Errorf("   got index %d", got)
		t.Errorf("  want index %d", want)
	}

	err := fmt.Errorf("kaboom at layer %d", 4711)
	fields, index = r.With("error", err, "k1", "v1", "k2", "v2").Fields()
	if got, want := len(fields), 3; got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	} else if got = len(index); got != want {
		t.Errorf("   got index %d", got)
		t.Errorf("  want index %d", want)
	}
}

func TestRecorderWithDanglingString(t *testing.T) {
	r := &Recorder{}

	fields, index := r.With("error").Fields()
	if got, want := len(fields), 1; got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	} else if got = len(index); got != want {
		t.Errorf("   got index %d", got)
		t.Errorf("  want index %d", want)
	}

	k, v := "error", "nil"
	if got, ok := fields[k]; !ok {
		t.Errorf("Missing key %q", k)
	} else if v != got {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", v)
	}
}

func TestRecorderWithDanglingNotString(t *testing.T) {
	r := &Recorder{}

	fields, index := r.With(1, 2, 3).Fields()
	if got, want := len(fields), 3; got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	} else if got = len(index); got != want {
		t.Errorf("   got index %d", got)
		t.Errorf("  want index %d", want)
	}

	k, v := "key0", 1
	if got, ok := fields[k]; !ok {
		t.Errorf("Missing key %q", k)
	} else if v != got {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", v)
	}

	k, v = "key2", 3
	if got, ok := fields[k]; !ok {
		t.Errorf("Missing key %q", k)
	} else if v != got {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", v)
	}
}

func TestRecorderWithFieldsError(t *testing.T) {
	r := &Recorder{}
	err := fmt.Errorf("kaboom at layer %d", 4711)
	fields, _ := r.With("error", err).Fields()
	if got, ok := fields["error"]; !ok {
		t.Errorf("With(\"error\", err) should not be ignored")
	} else if got != err.Error() {
		t.Errorf("With(\"%s\", \"%s\"):", "error", err)
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", err)
	}
}

func TestRecorderWithFieldsFunc(t *testing.T) {
	fn := func() string { return "Hello, world!" }

	r := &Recorder{}
	fields, _ := r.With("fn()", fn).Fields()
	if got, ok := fields["fn()"]; !ok {
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
	if got := len(r.With().Values); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	if got := len(r.WithMore().Values); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	err := fmt.Errorf("kaboom at layer %d", 4711)
	want = 5
	if got := len(r.With("error", err, "v1", "v2", "v3").Values); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestRecorderWithValuesFunc(t *testing.T) {
	fn := func() string { return "Hello, world!" }

	r := &Recorder{}
	values := r.With(fn).Values
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
	values := r.With(fn()).Values
	if len(values) <= 1 {
		t.Errorf("With(\"fn()\", fn) should not be ignored")
	} else if got, ok := values[1].(string); !ok {
		t.Errorf("Not string")
	} else if got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}
