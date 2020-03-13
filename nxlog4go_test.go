// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
	"github.com/ccpaging/nxlog4go/rolling"
)

const testFormat = "[%D %T %Z] [%L] (%S:%N) %M"
const testBenchFormat = "[%D %T] [%L] %M"
const testLogFile = "_logtest.log"
const benchLogFile = "_benchlog.log"

var now = time.Unix(0, 1234567890123456789).In(time.UTC)

func newRecorder(level int, prefix, src string, msg string) *driver.Recorder {
	return &driver.Recorder{
		Level:   level,
		Source:  src,
		Prefix:  prefix,
		Created: now,
		Message: msg,
	}
}

func TestELog(t *testing.T) {
	fmt.Printf("Testing %s\n", Version)
	lr := newRecorder(CRITICAL, "prefix", "source", "message")
	if lr.Level != CRITICAL {
		t.Errorf("Incorrect level: %d should be %d", lr.Level, CRITICAL)
	}
	if lr.Prefix != "prefix" {
		t.Errorf("Incorrect prefix: %s should be %s", lr.Prefix, "prefix")
	}
	if lr.Source != "source" {
		t.Errorf("Incorrect source: %s should be %s", lr.Source, "source")
	}
	if lr.Message != "message" {
		t.Errorf("Incorrect message: %s should be %s", lr.Source, "message")
	}
}

var recorderWriteTests = []struct {
	Test    string
	Record  *driver.Recorder
	Console string
}{
	{
		Test: "Normal message",
		Record: &driver.Recorder{
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

	layout := patt.NewLayout(testFormat, "utc", true)
	for _, test := range recorderWriteTests {
		name := test.Test

		// Pipe write and read must be in diff routines otherwise cause dead lock
		go func() {
			out := new(bytes.Buffer)
			layout.Encode(out, test.Record)
			w.Write(out.Bytes())
		}()

		n, _ := r.Read(buf)
		if got, want := string(buf[:n]), test.Console; got != (want + "\n") {
			t.Errorf("%s - %q:", name, testFormat)
			t.Errorf("%s:  got %q", name, got)
			t.Errorf("%s: want %q", name, want)
		}
	}
}

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)

	l := NewLogger(WARN).SetOutput(buf).SetOptions("format", "[%L] (%S) %M")

	if l == nil {
		t.Fatalf("NewLogger should never return nil")
	}
	if l.stdf.level != WARN {
		t.Fatalf("NewLogger produced invalid logger (incorrect level)")
	}

	if l.enabled(DEBUG) {
		t.Fatalf("NewLogger produced invalid enabled()")
	}

	//func (l *Logger) Warn(args ...interface{}) error {}
	if err := l.Warn("%s %d %#v", "Warn:", 1, []int{}); err.Error() != "Warn: 1 []int{}" {
		t.Errorf("Warn returned invalid error: %s", err)
	}
	want := "[WARN] (nxlog4go/nxlog4go_test.go) Warn: 1 []int{}\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Error(args ...interface{}) error {}
	if err := l.Error("%s %d %#v", "Error:", 10, []string{}); err.Error() != "Error: 10 []string{}" {
		t.Errorf("Error returned invalid error: %s", err)
	}
	want = "[EROR] (nxlog4go/nxlog4go_test.go) Error: 10 []string{}\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Critical(args ...interface{}) error {}
	if err := l.Critical("%s %d %#v", "Critical:", 100, []int64{}); err.Error() != "Critical: 100 []int64{}" {
		t.Errorf("Critical returned invalid error: %s", err)
	}
	want = "[CRIT] (nxlog4go/nxlog4go_test.go) Critical: 100 []int64{}\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	//func (l *Logger) Log(level int, args ...interface{}) {}
	l.Log(1, ERROR, "%s %d %#v", "Log Error:", 10, []string{})
	want = "[EROR] (nxlog4go/nxlog4go_test.go) Log Error: 10 []string{}\n"
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

func TestFileWriter(t *testing.T) {
	buf := new(bytes.Buffer)

	const (
		expected = "e7927ba6dc08038cf8ab631575169abf"
	)

	fw, err := os.Create(testLogFile)
	if err != nil {
		t.Error(err)
	}
	ww := io.MultiWriter(buf, fw)
	l := NewLogger(FINEST).SetOutput(ww).SetOptions("format", "[%L] %M")

	defer os.Remove(testLogFile)

	// Send some log messages
	l.Trace("This message is level %d", int(TRACE))
	want := "[TRAC] This message is level 3\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Debug("This message is level %s", Level(DEBUG))
	want = "[DEBG] This message is level DEBG\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Fine(func() string { return fmt.Sprintf("This message is level %v", Level(FINE)) })
	want = "[FINE] This message is level FINE\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Finest("This message is level %v", Level(FINEST))
	want = "[FNST] This message is level FNST\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	l.Finest(Level(FINEST), "is also this message's level")
	want = "[FNST] FNST is also this message's level\n"
	if got := buf.String(); got != want {
		t.Errorf("   got %q", got)
		t.Errorf("  want %q", want)
	}
	buf.Reset()

	fw.Close()

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
	sl := NewLogger(INFO).SetOutput(new(bytes.Buffer)).SetOptions("format", testFormat)

	mallocs := 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Log(1, WARN, "This is a WARN message")
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per sl.Log(1, (WARN, \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Warn("%s is a log message with level %d", "This", WARN)
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per sl.Warn(WARN, \"%%s is a log message with level %%d\", \"This\", WARN): %d\n", mallocs/N)

	// Console logger (not logged)
	sl = NewLogger(INFO).SetOutput(os.Stderr).SetOptions("format", testFormat)

	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Debug("This is a DEBUG log message")
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per unlogged sl.Debug(WARN, \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted (not logged)
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Debug("%s is a log message with level %d", "This", DEBUG)
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per unlogged sl.Debug(WARN, \"%%s is a log message with level %%d\", \"This\", WARN): %d\n", mallocs/N)
}

func BenchmarkConsoleWithCallerWriter(b *testing.B) {
	sl := NewLogger(INFO).SetOutput(ioutil.Discard).SetOptions("format", testBenchFormat)

	for i := 0; i < b.N; i++ {
		sl.Log(1, WARN, "This is a log message")
	}
}

func BenchmarkConsoleWriter(b *testing.B) {
	sl := NewLogger(INFO).SetOutput(ioutil.Discard).SetOptions("format", testBenchFormat, "caller", false)

	for i := 0; i < b.N; i++ {
		sl.Log(1, WARN, "This is a log message")
	}
}

func BenchmarkConsoleUtilWriter(b *testing.B) {
	sl := NewLogger(INFO).SetOutput(ioutil.Discard).SetOptions("format", testBenchFormat, "caller", false)

	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
}

func BenchmarkConsoleUtilNotWriter(b *testing.B) {
	sl := NewLogger(INFO).SetOutput(ioutil.Discard).SetOptions("format", testBenchFormat, "caller", false)

	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
}

func BenchmarkFileWriter(b *testing.B) {
	w, err := os.Create(benchLogFile)
	if err != nil {
		b.Error(err)
	}
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()

	sl := NewLogger(INFO).SetOutput(w).SetOptions("format", testBenchFormat, "caller", false)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(1, WARN, "This is a log message")
	}
	b.StopTimer()
}

func BenchmarkFileUtilWriter(b *testing.B) {
	w, err := os.Create(benchLogFile)
	if err != nil {
		b.Error(err)
	}
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()

	sl := NewLogger(INFO).SetOutput(w).SetOptions("format", testBenchFormat, "caller", false)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
}

func BenchmarkFileBufWriter(b *testing.B) {
	w := rolling.NewWriter(benchLogFile, 0)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()

	sl := NewLogger(INFO).SetOutput(w).SetOptions("format", testBenchFormat, "caller", false)
	// create file before benchmark testing
	sl.Log(1, WARN, "This is a log message")

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(1, WARN, "This is a log message")
	}
	b.StopTimer()
}

func BenchmarkFileBufUtilWriter(b *testing.B) {
	w := rolling.NewWriter(benchLogFile, 0)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	b.StopTimer()

	sl := NewLogger(INFO).SetOutput(w).SetOptions("format", testBenchFormat, "caller", false)
	// create file before benchmark testing
	sl.Info("%s is a log message", "This")

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
}

/*
go test -bench=.
goos: windows
goarch: amd64
pkg: github.com/ccpaging/nxlog4go
BenchmarkPrintln-4                       1339130               883 ns/op
BenchmarkPrintlnNoFlags-4                1719094               696 ns/op
BenchmarkConsoleWithCallerWriter-4        600030              1985 ns/op
BenchmarkConsoleWriter-4                 1520822               780 ns/op
BenchmarkConsoleUtilWriter-4             1246032               971 ns/op
BenchmarkConsoleUtilNotWriter-4         141834582                8.43 ns/op
BenchmarkFileWriter-4                     199998              5560 ns/op
BenchmarkFileUtilWriter-4                 181804              6397 ns/op
BenchmarkFileBufWriter-4                 1000000              1042 ns/op
BenchmarkFileBufUtilWriter-4              999883              1309 ns/op
*/
