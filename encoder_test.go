// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"testing"
	"time"
)

func TestLevelEncoder(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  string
	}{
		{"", DEBUG, "DEBG"},
		{"upper", TRACE, "TRACE"},
		{"upperColor", INFO, "\x1b[37mINFO\x1b[0m"},
		{"lower", WARN, "warn"},
		{"lowerColor", CRITICAL, "\x1b[31;1mcritical\x1b[0m"},
		{"std", CRITICAL + 1, "Level(8)"},
	}

	var out bytes.Buffer
	for _, tt := range tests {
		enc := NewLevelEncoder(tt.name)
		enc(&out, tt.level)
		if got := string(out.Bytes()); got != tt.want {
			t.Errorf("Incorrect level format of [%s]: %q should be %q", tt.name, got, tt.want)
		}
		out.Reset()
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
	var enc CallerEncoder
	for _, tt := range tests {
		enc = NewCallerEncoder(tt.name)
		enc(&out, filename)
		if got := string(out.Bytes()); got != tt.want {
			t.Errorf("Incorrect caller format of [%s]: %q should be %q", tt.name, got, tt.want)
		}
		out.Reset()
	}
}

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
		for _, tt := range tests {
			enc := NewDateEncoder(tt.name)
			enc(&out, &now)
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
		for _, tt := range tests {
			enc := NewTimeEncoder(tt.name)
			enc(&out, &now)
			want := now.Format(tt.format)
			if got := string(out.Bytes()); got != want {
				t.Errorf("Incorrect time format of [%s]: %q should be %q", tt.name, got, want)
			}
			out.Reset()
		}
	}
}

func TestFieldsEncoder(t *testing.T) {
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
	enc := NewFieldsEncoder("std")
	enc(out, data, index)

	want := " int=3 short=abcdefghijk long=0123456789abcdefg"
	if got := out.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}
