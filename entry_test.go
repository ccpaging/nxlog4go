package nxlog4go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

func TestEntryWithFields(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLogger(FINEST).SetOptions("format", "%L %S %M.%F").SetOutput(buf)
	e := NewEntry(l).With()
	want := 0
	if got := len(e.rec.Fields); got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	} else if got = len(e.rec.Index); got != want {
		t.Errorf("   got index %d", got)
		t.Errorf("  want index %d", want)
	}

	e = NewEntry(l).With("k1", "v1", "k2", "v2")
	want = 2
	if got := len(e.rec.Fields); got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	} else if got = len(e.rec.Index); got != want {
		t.Errorf("   got index %d", got)
		t.Errorf("  want index %d", want)
	}
}

func TestEntryFormatKeyValue(t *testing.T) {
	errBoom := fmt.Errorf("boom time")

	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOutput(buf).SetOptions("format", "%L %S %M.%F", "fieldsEncoder", "quote")
	e := NewEntry(l)

	e.With("err", errBoom, "k2", "v2", "k1", "v1").Log(1, ERROR, "kaboom")
	want := "EROR nxlog4go/entry_test.go kaboom. err=\"boom time\" k2=\"v2\" k1=\"v1\"\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestEntryFieldsJson(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOutput(buf).SetLayout(patt.NewJSONLayout("callerEncoder", "fullpath"))
	e := NewEntry(l).SetPrefix("test").With("source", "TestEntryFieldsJson")

	now := time.Now()
	errBoom := fmt.Errorf("boom")
	e.Log(1, ERROR, "kaboom", "error", errBoom)

	r := e.rec
	b := buf.Bytes()
	um := &driver.Recorder{}

	if err := json.Unmarshal(b, &um); err != nil {
		t.Errorf("   got %q", b)
		t.Errorf(" error %v", err)
	}

	if got, want := um.Level, ERROR; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Created, now; got.Unix() < want.Unix() {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Prefix, "test"; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Source, "entry_test.go"; strings.HasSuffix(got, want) == false {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Line, 61; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Message, "kaboom"; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if len(um.Fields) < 1 {
		t.Errorf("   got %s", b)
		t.Errorf("Missing Fields %v", um.Fields)
	}

	if want, ok := r.Fields["source"]; !ok {
		t.Errorf("Missing want field %q", "source")
	} else if wantStr, ok := want.(string); !ok {
		t.Errorf("Missing want type [%T]", want)
	} else if got, ok := um.Fields["source"]; !ok {
		t.Errorf("Missing got field %q", "source")
	} else if gotStr, ok := got.(string); !ok {
		t.Errorf("Missing got type [%T]", got)
	} else if gotStr != wantStr {
		t.Errorf("   got %q", gotStr)
		t.Errorf("  want %q", wantStr)
	}
}

func TestEntryWithValues(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLogger(FINEST).SetOptions("fields", false, "format", "%L %S %M.%F").SetOutput(buf)
	e := NewEntry(l).With()
	want := 0
	if got := len(e.rec.Values); got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	}

	e = NewEntry(l).With("v1", "v2", "k3")
	want = 3
	if got := len(e.rec.Values); got != want {
		t.Errorf("   got %d", got)
		t.Errorf("  want %d", want)
	}
}

func TestEntryValuesJson(t *testing.T) {
	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOptions("fields", false).SetOutput(buf).SetLayout(patt.NewJSONLayout("callerEncoder", "fullpath"))
	e := NewEntry(l).SetPrefix("test").With("source", "TestEntryValuesJson")

	errBoom := fmt.Errorf("boom")
	e.Log(1, ERROR, "kaboom", "error", errBoom)

	r := e.rec
	b := buf.Bytes()
	um := &driver.Recorder{}

	if err := json.Unmarshal(b, &um); err != nil {
		t.Errorf("   got %q", b)
		t.Errorf(" error %v", err)
	}

	if got, want := um.Prefix, "test"; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := um.Message, "kaboom"; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if len(um.Values) != 4 {
		t.Errorf("   got %s", b)
		t.Errorf("Missing Values %v", um.Values)
	}

	v, v1 := r.Values[0], um.Values[0]
	if wantStr, ok := v.(string); !ok {
		t.Errorf("Missing want type [%T]", v)
	} else if gotStr, ok := v1.(string); !ok {
		t.Errorf("Missing got type [%T]", v1)
	} else if gotStr != wantStr {
		t.Errorf("   got %q", gotStr)
		t.Errorf("  want %q", wantStr)
	}

	v = um.Values[3]
	if gotStr, ok := v.(string); !ok {
		t.Errorf("Missing got type [%T]", v)
	} else if wantStr := "boom"; gotStr != wantStr {
		t.Errorf("   got %q", gotStr)
		t.Errorf("  want %q", wantStr)
	}
}
