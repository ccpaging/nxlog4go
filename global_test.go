// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"testing"
)

const testGlobalPattern = "[%L] (%s) %M"

func TestGlobal(t *testing.T) {
	buf := new(bytes.Buffer)

	l := GetLogger().SetOutput(buf).Set("level", WARN).Set("pattern", testGlobalPattern)
	if l == nil {
		t.Fatalf("New should never return nil")
	}
	if l.level != WARN {
		t.Fatalf("New produced invalid logger (incorrect level)")
	}

	//func (l *Logger) Warn(args ...interface{}) error {}
	if err := Warn("%s %d %#v", "Warning:", 1, []int{}); err.Error() != "Warning: 1 []int{}" {
		t.Errorf("Warn returned invalid error: %s", err)
	}
	want := "[WARN] (global_test.go) Warning: 1 []int{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Error(args ...interface{}) error {}
	if err := Error("%s %d %#v", "Error:", 10, []string{}); err.Error() != "Error: 10 []string{}" {
		t.Errorf("Error returned invalid error: %s", err)
	}
	want = "[EROR] (global_test.go) Error: 10 []string{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Critical(args ...interface{}) error {}
	if err := Critical("%s %d %#v", "Critical:", 100, []int64{}); err.Error() != "Critical: 100 []int64{}" {
		t.Errorf("Critical returned invalid error: %s", err)
	}
	want = "[CRIT] (global_test.go) Critical: 100 []int64{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
}
