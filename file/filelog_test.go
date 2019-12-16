// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
)

const testLogFile = "_logtest.log"
const oldfiles = "_logtest.*.log"

const benchLogFile = "_benchlog.log"

func init() {
	os.Remove(testLogFile)
}

func newLogRecord(level int, src string, msg string) *l4g.Recorder {
	return &l4g.Recorder{
		Level:   level,
		Source:  src,
		Created: time.Now(),
		Message: msg,
	}
}

func removeFile(t *testing.T, filename string) {
	err := os.Remove(filename)
	if err != nil && t != nil {
		t.Errorf("remove (%q): %s", filename, err)
	}
}

func TestFileAppender(t *testing.T) {
	w, _ := NewFileAppender(testLogFile)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer removeFile(t, testLogFile)

	w.Enabled(newLogRecord(l4g.CRITICAL, "source", "message"))
	w.Close()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 52 {
		t.Errorf("malformed FileAppender: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestFileAppenderNotLogged(t *testing.T) {
	w, _ := NewFileAppender(testLogFile, "level", l4g.INFO)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}

	w.Enabled(newLogRecord(l4g.DEBUG, "source", "message"))
	w.Close()

	if contents, err := ioutil.ReadFile(testLogFile); err == nil {
		t.Errorf("malformed FileAppender: %q (%d bytes)", string(contents), len(contents))
	}
}

func writeSomethingToLogFile(log *l4g.Logger) {
	log.Finest("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Error("Time to close out!")
}

func TestFileLog(t *testing.T) {
	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	// Create a default logger that is logging messages of FINE or higher
	a, _ := NewFileAppender(testLogFile, "rotate", 0)
	f := l4g.NewFilter(l4g.FINE, nil, a)
	log.Attach(f)
	writeSomethingToLogFile(log)
	log.Detach(f)
	f.Close()
	defer removeFile(t, testLogFile)

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 178 {
		t.Errorf("malformed FileLog: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestFileLogRotate(t *testing.T) {
	// Enable internal logger
	l4g.GetLogLog().SetOptions("level", l4g.TRACE, "caller", true, "format", "[%D %T] [%L] (%S:%N) \t%M")
	defer l4g.GetLogLog().SetOptions("level", l4g.CRITICAL+1)

	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	// Can also specify manually via the following: (these are the defaults)
	a, _ := NewFileAppender(testLogFile,
		"level", l4g.FINE,
		"format", "[%D %T] [%L] (%S:%N) %M",
		"rotate", 10,
		"cycle", 5,
		"maxsize", "5k")

	log.Attach(l4g.NewFilter(l4g.FINE, nil, a))
	// Log some experimental messages
	for j := 0; j < 15; j++ {
		for i := 0; i < 200/(j+1); i++ {
			writeSomethingToLogFile(log)
		}
		time.Sleep(1 * time.Second)
	}
	log.Close()

	os.Remove(testLogFile)

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(oldfiles)
	fmt.Printf("%d files match %s\n", len(files), oldfiles)
	if len(files) != 3 {
		t.Errorf("FileRotateLog create %d files which should be 3", len(files))
	}
	for _, fname := range files {
		removeFile(t, fname)
	}
}

func BenchmarkCacheFileLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOptions("caller", false).SetOutput(nil)
	b.StopTimer()
	a, _ := NewFileAppender(benchLogFile,
		"level", l4g.INFO, "rotate", 0)
	sl.Attach(l4g.NewFilter(l4g.INFO, nil, a))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Warn("This is a log message")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileNotLogged(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOptions("caller", false).SetOutput(nil)
	b.StopTimer()
	a, _ := NewFileAppender(benchLogFile,
		"level", l4g.INFO, "rotate", 0)
	sl.Attach(l4g.NewFilter(l4g.INFO, nil, a))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("This is a log message")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileUtilLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOptions("caller", false).SetOutput(nil)
	b.StopTimer()
	a, _ := NewFileAppender(benchLogFile,
		"level", l4g.INFO, "rotate", 0)
	sl.Attach(l4g.NewFilter(l4g.INFO, nil, a))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileUtilNotLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOptions("caller", false).SetOutput(nil)
	b.StopTimer()
	a, _ := NewFileAppender(benchLogFile,
		"level", l4g.INFO, "rotate", 0)
	sl.Attach(l4g.NewFilter(l4g.WARN, nil, a))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	sl.Close()
	a.Close()
	os.Remove(benchLogFile)
}

/*
goos: windows
goarch: amd64
pkg: github.com/ccpaging/nxlog4go/file
BenchmarkCacheFileLog-4                   571123              1846 ns/op
BenchmarkCacheFileNotLogged-4            5239846               225 ns/op
BenchmarkCacheFileUtilLog-4               666751              1776 ns/op
BenchmarkCacheFileUtilNotLog-4           2850147               431 ns/op
*/
