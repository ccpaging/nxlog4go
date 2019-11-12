// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"testing"
)

const testLogLogPattern = "%P %L %M"

func TestLogLogger(t *testing.T) {
	buf := new(bytes.Buffer)

	l := GetLogLog().SetOutput(buf).Set("level", TRACE).Set("pattern", testLogLogPattern)
	if l == nil {
		t.Fatalf("GetLogLog should never return nil")
	}
	if l.level != TRACE {
		t.Fatalf("New produced invalid logger (incorrect level)")
	}

	//func (l *Logger) Warn(args ...interface{}) error {}
	if err := l.Warn("%s %d %#v", "Warn:", 1, []int{}); err.Error() != "Warn: 1 []int{}" {
		t.Errorf("Warn returned invalid error: %s", err)
	}
	want := "logg WARN Warn: 1 []int{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Error(args ...interface{}) error {}
	if err := l.Error("%s %d %#v", "Error:", 10, []string{}); err.Error() != "Error: 10 []string{}" {
		t.Errorf("Error returned invalid error: %s", err)
	}
	want = "logg EROR Error: 10 []string{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}
