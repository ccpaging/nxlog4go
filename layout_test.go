// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

var formatTests = []struct {
	Test    string
	Record  *Entry
	Formats map[string]string
	Args    map[string][]interface{}
}{
	{
		Test: "Standard formats",
		Record: &Entry{
			Level:   ERROR,
			Source:  "source",
			Message: "message",
			Created: now,
		},
		Formats: map[string]string{
			// TODO(kevlar): How can I do this so it'll work outside of PST?
			FormatDefault: "[2009/02/13 23:31:30 UTC] [EROR] (source:0) message\n",
			FormatShort:   "[23:31 13/02/09] [EROR] message\n",
			FormatAbbrev:  "[EROR] message\n",
		},
		Args: map[string][]interface{}{
			FormatShort: []interface{}{"timeEncoder", "hhmm", "dateEncoder", "dmy"},
		},
	},
}

func TestPatternLayout(t *testing.T) {
	out := new(bytes.Buffer)
	for _, test := range formatTests {
		name := test.Test
		for format, want := range test.Formats {
			layout := NewPatternLayout(format, "utc", true)
			if args, ok := test.Args[format]; ok {
				layout.Set(args...)
			}
			layout.Encode(out, test.Record)
			if got := out.String(); got != want {
				t.Errorf("%s - %q:", name, format)
				t.Errorf("   got %q", got)
				t.Errorf("  want %q", want)
			}
			out.Reset()
		}
	}
}

func TestKeyValueEncoder(t *testing.T) {
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

	out := new(bytes.Buffer)
	keyvalFieldsEncoder(out, data, index)

	want := " int=3 short=abcdefghijk long=\"0123456789abcdefg\""
	if got := out.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func BenchmarkPatternLayout(b *testing.B) {
	const updateEvery = 1
	e := &Entry{
		Level:   CRITICAL,
		Created: now,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	layout := NewPatternLayout(testFormat)
	out := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		e.Created = e.Created.Add(1 * time.Second / updateEvery)
		layout.Encode(out, e)
		out.Reset()
	}
}

func BenchmarkJson(b *testing.B) {
	const updateEvery = 1
	e := &Entry{
		Level:   CRITICAL,
		Created: now,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	for i := 0; i < b.N; i++ {
		e.Created = e.Created.Add(1 * time.Second / updateEvery)
		json.Marshal(e)
	}
}

func BenchmarkJsonLayout(b *testing.B) {
	const updateEvery = 1
	e := &Entry{
		Level:   CRITICAL,
		Created: now,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	layout := NewJSONLayout()
	out := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		e.Created = e.Created.Add(1 * time.Second / updateEvery)
		layout.Encode(out, e)
		out.Reset()
	}
}
