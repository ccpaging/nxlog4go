// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package patt

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/ccpaging/logger/driver"
)

var created = time.Unix(0, 1234567890123456789).In(time.UTC)

var formatTests = []struct {
	Test    string
	Record  *driver.Recorder
	Formats map[string]string
	Args    map[string][]interface{}
}{
	{
		Test: "Standard formats",
		Record: &driver.Recorder{
			Level:   0,
			Source:  "source",
			Message: "message",
			Created: created,
		},
		Formats: map[string]string{
			// TODO(kevlar): How can I do this so it'll work outside of PST?
			FormatDefault: "[2009/02/13 23:31:30 UTC] [] (source:0) message\n",
			FormatShort:   "[23:31 13/02/09] [] message\n",
			FormatAbbrev:  "[] message\n",
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
			layout := NewLayout(format, "utc", true)
			if args, ok := test.Args[format]; ok {
				layout.SetOptions(args...)
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

func BenchmarkPatternLayout(b *testing.B) {
	const updateEvery = 1
	r := &driver.Recorder{
		Level:   0,
		Created: created,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	layout := NewLayout("[%D %T %Z] [%L] (%S:%N) %M")
	out := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		r.Created = r.Created.Add(1 * time.Second / updateEvery)
		layout.Encode(out, r)
		out.Reset()
	}
}

func BenchmarkJson(b *testing.B) {
	const updateEvery = 1
	r := &driver.Recorder{
		Level:   0,
		Created: created,
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
	r := &driver.Recorder{
		Level:   0,
		Created: created,
		Prefix:  "prefix",
		Source:  "source",
		Message: "message",
	}
	layout := NewJSONLayout(nil)
	out := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		r.Created = r.Created.Add(1 * time.Second / updateEvery)
		layout.Encode(out, r)
		out.Reset()
	}
}
