// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package nxlog4go

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"
	"bytes"
)

const testPattern = "[%D %T %z] [%L] (%s:%N) %M\n"
const testLogFile = "_logtest.log"
const benchLogFile = "_benchlog.log"

var now = time.Unix(0, 1234567890123456789).In(time.UTC)

var logRecordWriteTests = []struct {
	Test    string
	Record  *LogRecord
	Console string
}{
	{
		Test: "Normal message",
		Record: &LogRecord{
			Level:   CRITICAL,
			Source:  "source",
			Message: "message",
			Created: now,
		},
		Console: "[2009/02/13 23:31:30 UTC] [CRIT] (source:0) message",
	},
}

func TestConsoleWriter(t *testing.T) {
	r, w := io.Pipe()

	buf := make([]byte, 1024)

	layout := NewPatternLayout(testPattern).Set("utc", true)
	for _, test := range logRecordWriteTests {
		name := test.Test

		// Pipe write and read must be in diff routines otherwise cause dead lock
		go w.Write(layout.Format(test.Record))

		n, _ := r.Read(buf)
		if got, want := string(buf[:n]), test.Console; got != (want + "\n") {
			t.Errorf("%s:  got %q", name, got)
			t.Errorf("%s: want %q", name, want)
		}
	}
}

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)

	l := NewLogger(WARNING).SetOutput(buf).Set("pattern", "[%L] (%s) %M")

	if l == nil {
		t.Fatalf("New should never return nil")
	}
	if l.level != WARNING {
		t.Fatalf("New produced invalid logger (incorrect level)")
	}

	//func (l *Logger) Warn(args ...interface{}) error {}
	if err := l.Warn("%s %d %#v", "Warning:", 1, []int{}); err.Error() != "Warning: 1 []int{}" {
		t.Errorf("Warn returned invalid error: %s", err)
	}
	want := "[WARN] (nxlog4go_test.go) Warning: 1 []int{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Error(args ...interface{}) error {}
	if err := l.Error("%s %d %#v", "Error:", 10, []string{}); err.Error() != "Error: 10 []string{}" {
		t.Errorf("Error returned invalid error: %s", err)
	}
	want = "[EROR] (nxlog4go_test.go) Error: 10 []string{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Critical(args ...interface{}) error {}
	if err := l.Critical("%s %d %#v", "Critical:", 100, []int64{}); err.Error() != "Critical: 100 []int64{}" {
		t.Errorf("Critical returned invalid error: %s", err)
	}
	want = "[CRIT] (nxlog4go_test.go) Critical: 100 []int64{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Log(level int, args ...interface{}) {}
	l.Log(ERROR, "%s %d %#v", "Log Error:", 10, []string{})
	want = "[EROR] (nxlog4go_test.go) Log Error: 10 []string{}"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	// Already tested or basically untestable
	//func (l *Logger) Finest(args ...interface{}) {}
	//func (l *Logger) Fine(args ...interface{}) {}
	//func (l *Logger) Debug(args ...interface{}) {}
	//func (l *Logger) Trace(args ...interface{}) {}
	//func (l *Logger) Info(args ...interface{}) {}
}

func TestLogOutput(t *testing.T) {
	buf := new(bytes.Buffer)

	const (
		expected = "e7927ba6dc08038cf8ab631575169abf"
	)

	fbw := NewFileBufWriter(testLogFile).SetFlush(0)
	ww := io.MultiWriter(buf, fbw)
	l := &Logger{
		out:     ww,
		level:   FINEST,
		caller:  true,
		layout:  NewPatternLayout("[%L] %M\n"),
	}

	defer os.Remove(testLogFile)

	// Send some log messages
	l.Trace("This message is level %d", int(TRACE))
	want := "[TRAC] This message is level 3\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Debug("This message is level %s", DEBUG)
	want = "[DEBG] This message is level DEBG\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Fine(func() string { return fmt.Sprintf("This message is level %v", FINE) })
	want = "[FINE] This message is level FINE\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Finest("This message is level %v", FINEST)
	want = "[FNST] This message is level FNST\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Finest(FINEST, "is also this message's level")
	want = "[FNST] FNST is also this message's level\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	fbw.Close()

	contents, err := ioutil.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Could not read output log: %s", err)
	}

	sum := md5.New()
	sum.Write(contents)
	if sumstr := hex.EncodeToString(sum.Sum(nil)); sumstr != expected {
		t.Errorf("--- Log Contents:\n%s---", string(contents))
		t.Fatalf("Checksum does not match: %s (expecting %s)", sumstr, expected)
	}
}

func TestCountMallocs(t *testing.T) {
	const N = 1
	var m runtime.MemStats
	getMallocs := func() uint64 {
		runtime.ReadMemStats(&m)
		return m.Mallocs
	}

	// Console logger
	sl := &Logger{
		out:     new(bytes.Buffer),
		level:   INFO,
		caller:  true,
		layout:  NewPatternLayout(testPattern),
	}

	mallocs := 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.intLog(WARNING, "This is a WARNING message")
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per sl.intLog((WARNING, \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Warn("%s is a log message with level %d", "This", WARNING)
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per sl.Warn(WARNING, \"%%s is a log message with level %%d\", \"This\", WARNING): %d\n", mallocs/N)

	// Console logger (not logged)
	sl = &Logger{
		out:     os.Stderr,
		level:   INFO,
		caller:  true,
		layout:  NewPatternLayout(testPattern),
	}
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Debug("This is a DEBUG log message")
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per unlogged sl.Debug(WARNING, \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted (not logged)
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Debug("%s is a log message with level %d", "This", DEBUG)
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per unlogged sl.Debug(WARNING, \"%%s is a log message with level %%d\", \"This\", WARNING): %d\n", mallocs/N)
}

func BenchmarkConsoleWriter(b *testing.B) {
	sl := &Logger{
		out:     ioutil.Discard,
		level:   INFO,
		caller:  false,
		layout:  NewPatternLayout(testPattern),
	}
	for i := 0; i < b.N; i++ {
		sl.intLog(WARNING, "This is a log message")
	}
}

func BenchmarkConsoleUtilWriter(b *testing.B) {
	sl := &Logger{
		out:     ioutil.Discard,
		level:   INFO,
		caller:  true,
		layout:  NewPatternLayout(testPattern),
	}
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
}

func BenchmarkConsoleUtilNotWriter(b *testing.B) {
	sl := &Logger{
		out:     ioutil.Discard,
		level:   INFO,
		caller:  true,
		layout:  NewPatternLayout(testPattern),
	}
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
}

func BenchmarkFileWriter(b *testing.B) {
	w := NewFileBufWriter(benchLogFile).SetFlush(0)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	b.StopTimer()
	sl := &Logger{
		out:     w,
		level:   INFO,
		caller:  false,
		layout:  NewPatternLayout(testPattern),
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.intLog(WARNING, "This is a log message")
	}
	b.StopTimer()
}

func BenchmarkFileUtilWriter(b *testing.B) {
	w := NewFileBufWriter(benchLogFile).SetFlush(0)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	defer w.Close()
	b.StopTimer()
	sl := &Logger{
		out:     w,
		level:   INFO,
		caller:  false,
		layout:  NewPatternLayout(testPattern),
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
}

func BenchmarkFileBufWriter(b *testing.B) {
	w := NewFileBufWriter(benchLogFile)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	b.StopTimer()
	sl := &Logger{
		out:     w,
		level:   INFO,
		caller:  false,
		layout:  NewPatternLayout(testPattern),
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.intLog(WARNING, "This is a log message")
	}
	b.StopTimer()
}

func BenchmarkFileBufUtilWriter(b *testing.B) {
	w := NewFileBufWriter(benchLogFile)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	b.StopTimer()
	sl := &Logger{
		out:     w,
		level:   INFO,
		caller:  false,
		layout:  NewPatternLayout(testPattern),
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
}

// go test -bench=.
// goos: windows
// goarch: amd64
// pkg: github.com/ccpaging/nxlog4go
// BenchmarkItoa-4                         10000000               134 ns/op
// BenchmarkPrintln-4                       2000000               993 ns/op
// BenchmarkPrintlnNoFlags-4                2000000               772 ns/op
// BenchmarkPatternLayout-4                 2000000               656 ns/op
// BenchmarkJson-4                           500000              3000 ns/op
// BenchmarkJsonLayout-4                    1000000              1067 ns/op
// BenchmarkConsoleWriter-4                 1000000              1209 ns/op
// BenchmarkConsoleUtilWriter-4              500000              3658 ns/op
// BenchmarkConsoleUtilNotWriter-4          5000000               252 ns/op
// BenchmarkFileWriter-4                     200000              6740 ns/op
// BenchmarkFileUtilWriter-4                 200000              7465 ns/op
// BenchmarkFileBufWriter-4                 1000000              1681 ns/op
// BenchmarkFileBufUtilWriter-4             1000000              1934 ns/op
