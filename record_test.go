package nxlog4go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

const testRecordPatternText = "%L %s %M.%F"
const testRecordPatternJSON = "%L %s %M. %J"

func newLogRecord(level Level, prefix, src string, msg string) *LogRecord {
	return &LogRecord{
		Level:   level,
		Source:  src,
		Prefix:  prefix,
		Created: now,
		Message: msg,
	}
}

func TestLogRecord(t *testing.T) {
	r := newLogRecord(CRITICAL, "prefix", "source", "message")
	if r.Level != CRITICAL {
		t.Errorf("Incorrect level: %d should be %d", r.Level, CRITICAL)
	}
	if r.Prefix != "prefix" {
		t.Errorf("Incorrect prefix: %s should be %s", r.Prefix, "prefix")
	}
	if r.Source != "source" {
		t.Errorf("Incorrect source: %s should be %s", r.Source, "source")
	}
	if r.Message != "message" {
		t.Errorf("Incorrect message: %s should be %s", r.Source, "message")
	}
}

func TestLogRecordWithError(t *testing.T) {
	l := NewLogger(FINEST)
	r := NewLogRecord(l)
	err := fmt.Errorf("kaboom at layer %d", 4711)
	if got := r.With("error", err).Data["error"]; got != err.Error() {
		t.Errorf("With(\"%s\", \"%s\"):", "error", err)
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", err)
	}
}

func TestLogRecordWithArgs(t *testing.T) {
	l := NewLogger(FINEST)
	r := NewLogRecord(l)
	err := fmt.Errorf("kaboom at layer %d", 4711)
	want := 3
	if got := len(r.With("error", err, "k1", "v1", "k2", "v2").Data); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestLogRecordWithFunc(t *testing.T) {
	fn := func() string { return "Hello, world!" }

	l := NewLogger(FINEST)
	r := NewLogRecord(l)
	if got, ok := r.With("fn()", fn).Data["fn()"]; !ok {
		t.Errorf("With(\"fn()\", fn) should not be ignored")
	} else if s, ok := got.(string); !ok {
		t.Errorf("Not string")
	} else if s != fn() {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", fn())
	}
}

func TestRecordFormatText(t *testing.T) {
	errBoom := fmt.Errorf("boom time")

	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOutput(buf).Set("pattern", testRecordPatternText)
	r := NewLogRecord(l)

	r.With("err", errBoom).Log(ERROR, "kaboom")
	want := "EROR record_test.go kaboom. err=\"boom time\""
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestRecordFormatJson(t *testing.T) {
	errBoom := fmt.Errorf("boom time")

	buf := new(bytes.Buffer)
	l := NewLogger(INFO).SetOutput(buf).Set("pattern", PatternJSON)
	r := NewLogRecord(l)

	r.With("error", errBoom).Log(ERROR, "kaboom")

	b := buf.Bytes()
	r1 := NewLogRecord(l)

	if err := json.Unmarshal(b, &r1); err != nil {
		t.Errorf("   got %q", b)
		t.Errorf(" error %v", err)
	}

	if got, want := r1.Level, r.Level; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := r1.Created, r.Created; got.Unix() != want.Unix() {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := r1.Prefix, r.Prefix; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := r1.Source, r.Source; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := r1.Line, r.Line; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if got, want := r1.Message, r.Message; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}

	if len(r1.Data) < 1 {
		t.Errorf("Missing Data %v", r1.Data)
	} else if want, ok := r.Data["error"]; !ok {
		t.Errorf("Missing want field %q", "error")
	} else if wantStr, ok := want.(string); !ok {
		t.Errorf("Missing want type [%T]", want)
	} else if got, ok := r1.Data["error"]; !ok {
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
