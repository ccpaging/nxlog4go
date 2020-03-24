// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"testing"

	"github.com/ccpaging/nxlog4go/driver"
)

func TestAtAboveLevelEnabler(t *testing.T) {
	enb := driver.AtAbove(INFO)
	r, want := &driver.Recorder{Level: DEBUG}, false
	if got := enb.Enabled(r); got != want {
		t.Errorf("Got %v. Want %v", got, want)
	}

	r.Level, want = INFO, true
	if got := enb.Enabled(r); got != want {
		t.Errorf("Got %v. Want %v", got, want)
	}
}

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

	e := NewLevelEncoder("")

	out := new(bytes.Buffer)
	r := &driver.Recorder{}
	for _, tt := range tests {
		e = e.Open(tt.name)
		r.Level = tt.level
		e.Encode(out, r)
		if got := string(out.Bytes()); got != tt.want {
			t.Errorf("Incorrect level format of [%s]: %q should be %q", tt.name, got, tt.want)
		}
		out.Reset()
	}
}
