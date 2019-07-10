package nxlog4go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

const testEntryPatternText = "%L %s %M.%F"

func TestEntryWithError(t *testing.T) {
	l := NewLogger(FINEST)
	e := NewEntry(l)
	err := fmt.Errorf("kaboom at layer %d", 4711)
	if got := e.With("error", err).Data["error"]; got != err.Error() {
		t.Errorf("With(\"%s\", \"%s\"):", "error", err)
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", err)
	}
}

func TestEntryWithArgs(t *testing.T) {
	l := NewLogger(FINEST)
	e := NewEntry(l)
	err := fmt.Errorf("kaboom at layer %d", 4711)
	want := 3
	if got := len(e.With("error", err, "k1", "v1", "k2", "v2").Data); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	} else if got = len(e.index); got != want {
		t.Errorf("   got index %q", got)
		t.Errorf("  want index %q", want)
	}
}

func TestEntryWithFunc(t *testing.T) {
	fn := func() string { return "Hello, world!" }

	l := NewLogger(FINEST)
	e := NewEntry(l)
	if got, ok := e.With("fn()", fn).Data["fn()"]; !ok {
		t.Errorf("With(\"fn()\", fn) should not be ignored")
	} else if s, ok := got.(string); !ok {
		t.Errorf("Not string")
	} else if s != fn() {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", fn())
	}
}

func TestEntryFormatText(t *testing.T) {
	errBoom := fmt.Errorf("boom time")

	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOutput(buf).Set("pattern", testEntryPatternText)
	e := NewEntry(l)

	e.With("err", errBoom, "k2", "v2", "k1", "v1").Log(1, ERROR, "kaboom")
	want := "EROR entry_test.go kaboom. err=\"boom time\" k2=v2 k1=v1"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestEntryFormatJson(t *testing.T) {
	errBoom := fmt.Errorf("boom time")

	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOutput(buf).Set("pattern", PatternJSON)
	e := NewEntry(l)

	e.With("error", errBoom).Log(1, ERROR, "kaboom")

	b := buf.Bytes()
	um := NewEntry(l)

	if err := json.Unmarshal(b, &um); err != nil {
		t.Errorf("   got %q", b)
		t.Errorf(" error %v", err)
	}

	if got, want := um.Level, e.Level; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Created, e.Created; got.Unix() != want.Unix() {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Prefix, e.Prefix; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Source, e.Source; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Line, e.Line; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Message, e.Message; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if len(um.Data) < 1 {
		t.Errorf("Missing Data %v", um.Data)
	} else if want, ok := e.Data["error"]; !ok {
		t.Errorf("Missing want field %q", "error")
	} else if wantStr, ok := want.(string); !ok {
		t.Errorf("Missing want type [%T]", want)
	} else if got, ok := um.Data["error"]; !ok {
		t.Errorf("Missing got field %q", "error")
	} else if gotStr, ok := got.(string); !ok {
		t.Errorf("Missing got type [%T]", got)
	} else if gotStr != wantStr {
		t.Errorf("   got %q", gotStr)
		t.Errorf("  want %q", wantStr)
	} else {
		// Every thing is ok
	}

}
