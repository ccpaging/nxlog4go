// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"testing"
	"time"
	"encoding/json"
)

func TestFormatHMS(t *testing.T) {
	now := time.Now()
	//year, month, day := now.Date()
	out := bytes.NewBuffer(make([]byte, 0, 64))
	formatHMS(out, &now, ':')
	want := now.Format("15:04:05")
	if got := string(out.Bytes()); got != want {
		t.Errorf("Incorrect time format: %s should be %s", got, want)
	}
}

func TestFormatDMY(t *testing.T) {
	now := time.Now()
	out := bytes.NewBuffer(make([]byte, 0, 64))
	formatDMY(out, &now, '/')
	want := now.Format("02/01/06")
	if got := string(out.Bytes()); got != want {
		t.Errorf("Incorrect time format: %s should be %s", got, want)
	}
}

func TestFormatCYMD(t *testing.T) {
	now := time.Now()
	out := bytes.NewBuffer(make([]byte, 0, 64))
	formatCYMD(out, &now, '/')
	want := now.Format("2006/01/02")
	if got := string(out.Bytes()); got != want {
		t.Errorf("Incorrect time format: %s should be %s", got, want)
	}
}

var patternTests = []struct {
	Test     string
	Record   *LogRecord
	Patterns map[string]string
}{
	{
		Test: "Standard formats",
		Record: &LogRecord{
			Level:   ERROR,
			Source:  "source",
			Message: "message",
			Created: now,
		},
		Patterns: map[string]string{
			// TODO(kevlar): How can I do this so it'll work outside of PST?
			PatternDefault: "[2009/02/13 23:31:30 UTC] [EROR] (source:0) message\n",
			PatternShort:   "[23:31 13/02/09] [EROR] message\n",
			PatternAbbrev:  "[EROR] message\n",
		},
	},
}

func TestPatternLayout(t *testing.T) {
	for _, test := range patternTests {
		name := test.Test
		for pattern, want := range test.Patterns {
			layout := NewPatternLayout(pattern).Set("utc", true)
			if got := string(layout.Format(test.Record)); got != want {
				t.Errorf("%s - %s:", name, pattern)
				t.Errorf("   got %q", got)
				t.Errorf("  want %q", want)
			}
		}
	}
}

func TestPatternLayoutAndCaller(t *testing.T) {
	want := true
	if got := NewPatternLayout(PatternDefault).Caller(); got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}
	want = false
	if got := NewPatternLayout(PatternShort).Caller(); got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}
	want = false
	if got := NewPatternLayout(PatternAbbrev).Caller(); got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}
}

func TestFlagAndCallerSetting(t *testing.T) {
	var b bytes.Buffer
	want := false
	if got := New(&b, "Test:", LstdFlags).caller; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}
	want = true
	if got := New(&b, "Test:", LstdFlags | Lshortfile).caller; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}
	want = true
	if got := New(&b, "Test:", LstdFlags | Llongfile).caller; got != want {
		t.Errorf("   got %v", got)
		t.Errorf("  want %v", want)
	}
}

func TestDataField(t *testing.T) {
	buf := new(bytes.Buffer)

	data := map[string]interface{}{
		"int":   3,
		"short": "abcdefghijk",
		"long":  "0123456789abcdefg",
	}
	index := []string{
		"int",
		"short",
		"long",
	}

	for _, k := range index {
		writeKeyVal(buf, k, data[k])
	}

	want := " int=3 short=abcdefghijk long=\"0123456789abcdefg\""
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func BenchmarkPatternLayout(b *testing.B) {
	const updateEvery = 1
	r := &LogRecord{
		Level:   CRITICAL,
		Created: now,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	layout := NewPatternLayout(testPattern)
	for i := 0; i < b.N; i++ {
		r.Created = r.Created.Add(1 * time.Second / updateEvery)
		layout.Format(r)
	}
}

func BenchmarkJson(b *testing.B) {
	const updateEvery = 1
	r := &LogRecord{
		Level:   CRITICAL,
		Created: now,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	for i := 0; i < b.N; i++ {
		r.Created = r.Created.Add(1 * time.Second / updateEvery)
		json.Marshal(r)
	}
}

func BenchmarkJsonLayout(b *testing.B) {
	const updateEvery = 1
	r := &LogRecord{
		Level:   CRITICAL,
		Created: now,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	layout := NewPatternLayout(PatternJSON)
	for i := 0; i < b.N; i++ {
		r.Created = r.Created.Add(1 * time.Second / updateEvery)
		layout.Format(r)
	}
}
