package nxlog4go

import (
	"bytes"
	"log"
	"testing"
)

func TestNewStdLog(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := New(buf, "redirected", Lshortfile).SetOptions(
		"level", FINEST,
		"format", "[%L] [%P] (%S) %M%F")
	elog := logger.With("source", "testing")
	elog.Info("redirected.")

	want := "[INFO] [redirected] (stdlog_test.go) redirected. source=testing\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	log := NewStdLog(elog.AddCallerSkip(3))
	log.Println("This is stdlog's println.")
	want = "[INFO] [redirected] (stdlog_test.go) This is stdlog's println. source=testing\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	levels := []int{FINEST, DEBUG, TRACE, INFO, WARN, ERROR}
	for _, level := range levels {
		log := NewStdLogAt(elog.AddCallerSkip(3), level)
		log.Println("redirected.")

		want := "[" + Level(level).String() + "] [redirected] (stdlog_test.go) redirected. source=testing\n"
		if got := buf.String(); got != want {
			t.Errorf("   got %q", got)
			t.Errorf("  want %q", want)
		}
		buf.Reset()
	}
}

func TestRedirectStdLog(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetFlags(log.Lshortfile)
	log.SetOutput(buf)

	testStdlog := func() {
		log.Println("redirected")
		want := "stdlog_test.go:54: redirected\n"
		if got := buf.String(); got != want {
			t.Errorf("   got %q", got)
			t.Errorf("  want %q", want)
		}
		buf.Reset()
	}
	testStdlog()

	logger := New(buf, "redirected", Lshortfile).SetOptions(
		"level", FINEST,
		"format", "[%L] [%P] (%S) %M%F")
	elog := logger.With("source", "testing")

	restore := RedirectStdLogAt(elog, "debug")
	log.Println("redirected.")
	want := "[DEBG] [redirected] (stdlog_test.go) redirected. source=testing\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	restore()
	testStdlog()
}
