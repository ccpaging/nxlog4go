// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package patt

import (
	"bytes"
	"testing"
	"time"

	"github.com/ccpaging/nxlog4go/driver"
)

func TestDateEncoder(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		format string
	}{
		{"", "2006/01/02"},
		{"dmy", "02/01/06"},
		{"mdy", "01/02/06"},
		{"cymdDash", "2006-01-02"},
		{"cymdDot", "2006.01.02"},
		{"cymdSlash", "2006/01/02"},
	}

	var out bytes.Buffer
	for i := 0; i < 2; i++ {
		now = now.AddDate(0, 0, i)
		r := &driver.Recorder{Created: now}
		for _, tt := range tests {
			e := NewDateEncoder(tt.name)
			e.Encode(&out, r)
			want := now.Format(tt.format)
			if got := string(out.Bytes()); got != want {
				t.Errorf("Incorrect time format of [%s]: %q should be %q", tt.name, got, want)
			}
			out.Reset()
		}
	}
}

func TestTimeEncoder(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		format string
	}{
		{"", "15:04:05"},
		{"hhmm", "15:04"},
		{"hms", "15:04:05"},
		{"rfc3339nano", time.RFC3339Nano},
		{"iso8601", "2006-01-02T15:04:05.000Z0700"},
	}

	var out bytes.Buffer
	for i := 0; i < 2; i++ {
		now = now.Add(1 * time.Second)
		r := &driver.Recorder{Created: now}
		for _, tt := range tests {
			e := NewTimeEncoder(tt.name)
			e.Encode(&out, r)
			want := now.Format(tt.format)
			if got := string(out.Bytes()); got != want {
				t.Errorf("Incorrect time format of [%s]: %q should be %q", tt.name, got, want)
			}
			out.Reset()
		}
	}
}

func TestCallerEncoder(t *testing.T) {
	filename := "/home/jack/src/github.com/foo/foo.go"

	tests := []struct {
		name string
		want string
	}{
		{"", "foo/foo.go"},
		{"nopath", "foo.go"},
		{"fullpath", filename},
		{"shortpath", "foo/foo.go"},
	}

	var out bytes.Buffer
	r := &driver.Recorder{Source: filename}
	e0 := NewCallerEncoder("")
	for _, tt := range tests {
		e := e0.NewEncoder(tt.name)
		e.Encode(&out, r)
		if got := string(out.Bytes()); got != tt.want {
			t.Errorf("Incorrect caller format of [%s]: %q should be %q", tt.name, got, tt.want)
		}
		out.Reset()
	}
}

func TestFieldsEncoder(t *testing.T) {
	r := &driver.Recorder{
		Values: []interface{}{
			"int", 3,
			"short", "abcdefghijk",
			"long", "0123456789abcdefg",
		},
	}

	out := new(bytes.Buffer)
	e := NewFieldsEncoder("")
	e.Encode(out, r)

	want := " int=3 short=abcdefghijk long=0123456789abcdefg"
	if got := out.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func TestValuesEncoder(t *testing.T) {
	r := &driver.Recorder{
		Values: []interface{}{
			"int", 3,
			"short", "abcdefghijk",
			"long", "0123456789abcdefg",
		},
	}

	out := new(bytes.Buffer)
	e := NewValuesEncoder("")
	e.Encode(out, r)

	want := " int 3 short abcdefghijk long 0123456789abcdefg"
	if got := out.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}

func BenchmarkItoa(b *testing.B) {
	dst := make([]byte, 0, 64)
	for i := 0; i < b.N; i++ {
		dst = dst[0:0]
		itoa(&dst, 2015, 4)   // year
		itoa(&dst, 1, 2)      // month
		itoa(&dst, 30, 2)     // day
		itoa(&dst, 12, 2)     // hour
		itoa(&dst, 56, 2)     // minute
		itoa(&dst, 0, 2)      // second
		itoa(&dst, 987654, 6) // microsecond
	}
}

func BenchmarkRFC3339Nano(b *testing.B) {
	e := NewTimeEncoder("rfc3339nano")
	out := new(bytes.Buffer)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		r := &driver.Recorder{Created: time.Now()}
		e.Encode(out, r)
		out.Reset()
	}
	b.StopTimer()
}

func BenchmarkTimeFormat(b *testing.B) {
	out := new(bytes.Buffer)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		out.WriteString(time.Now().Format(time.RFC3339Nano))
		out.Reset()
	}
	b.StopTimer()
}
