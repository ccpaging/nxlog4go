// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package filelog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
)

const testLogFile = "_logtest.log"
const oldfiles = "_logtest.*.log"

const benchLogFile = "_benchlog.log"

var now = time.Unix(0, 1234567890123456789).In(time.UTC)

func newEntry(level int, src string, msg string) *l4g.Entry {
	return &l4g.Entry{
		Level:   level,
		Source:  src,
		Created: now,
		Message: msg,
	}
}

func TestFileAppender(t *testing.T) {
	w, _ := NewAppender(testLogFile, "rotate", false)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.Write(newEntry(l4g.CRITICAL, "source", "message"))
	runtime.Gosched()
	w.Close()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 52 {
		t.Errorf("malformed FileAppender: %q (%d bytes)", string(contents), len(contents))
	}
}

func writeSomethingToLogFile(log *l4g.Logger) {
	log.Finest("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Critical("Time to close out!")
}

func TestFileLog(t *testing.T) {
	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	// Create a default logger that is logging messages of FINE or higher
	a, _ := NewAppender(testLogFile, "rotate", false)
	filters := l4g.NewFilters().Add("file", l4g.FINE, a)
	log.SetFilters(filters)
	writeSomethingToLogFile(log)
	log.SetFilters(nil)
	filters.Close()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 168 {
		t.Errorf("malformed FileLog: %q (%d bytes)", string(contents), len(contents))
	}

	// Remove the file so it's not lying around
	err := os.Remove(testLogFile)
	if err != nil {
		t.Errorf("remove (%q): %s", testLogFile, err)
	}
}

func TestFileLogRotate(t *testing.T) {
	// Enable internal logger
	l4g.GetLogLog().Set("level", l4g.TRACE)
	defer l4g.GetLogLog().Set("level", l4g.Level(0).Max()+1)

	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	/* Can also specify manually via the following: (these are the defaults) */
	filter, _ := NewAppender(testLogFile, "rotate", true, "maxbackup", 10)
	filter.Set("format", "[%D %T] [%L] (%x) %M%R")
	filter.Set("cycle", 5).Set("clock", -1).Set("maxsize", "5k")

	filters := l4g.NewFilters().Add("file", l4g.FINE, filter)
	log.SetFilters(filters)

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		for i := 0; i < 200/(j+1); i++ {
			writeSomethingToLogFile(log)
		}
		time.Sleep(1 * time.Second)
	}
	// Close the log filters
	log.SetFilters(nil)
	// DO NOT FORGET CLOSING
	filters.Close()

	os.Remove(testLogFile)

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(oldfiles)
	fmt.Printf("%d files match %s\n", len(files), oldfiles)
	if len(files) != 3 {
		t.Errorf("FileRotateLog create %d files which should be 3", len(files))
	}
	for _, fname := range files {
		err := os.Remove(fname)
		if err != nil {
			t.Errorf("remove (%q): %s", fname, err)
		}
	}
}

func TestRotateFile(t *testing.T) {
	// Enable internal logger
	l4g.GetLogLog().Set("level", l4g.TRACE)
	defer l4g.GetLogLog().Set("level", l4g.Level(0).Max()+1)

	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	/* Can also specify manually via the following: (these are the defaults) */
	filter, _ := NewAppender(testLogFile, "rotate", true, "maxbackup", 10)
	filter.Set("format", "[%D %T] [%L] (%x) %M%R")
	filter.Set("cycle", 0).Set("maxsize", "5k")

	filters := l4g.NewFilters().Add("file", l4g.FINE, filter)
	log.SetFilters(filters)

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		time.Sleep(1 * time.Second)
		for i := 0; i < 200/(j+1); i++ {
			writeSomethingToLogFile(log)
		}
		//time.Sleep(1 * time.Second)
	}
	// Close the log filters
	log.SetFilters(nil)
	// DO NOT FORGET CLOSING
	filters.Close()

	os.Remove(testLogFile)

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(oldfiles)
	fmt.Printf("%d files match %s\n", len(files), oldfiles)
	if len(files) != 10 {
		t.Errorf("FileRotateLog create %d files which should be 10", len(files))
	}
	for _, fname := range files {
		err := os.Remove(fname)
		if err != nil {
			t.Errorf("remove (%q): %s", fname, err)
		}
	}
}

func BenchmarkFileLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", "flush", 0)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Warn("This is a log message")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkFileNotLogged(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 0)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("This is a log message")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkFileUtilLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 0)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkFileUtilNotLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 0)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 4096)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Warn("This is a log message")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileNotLogged(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 4096)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("This is a log message")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileUtilLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 4096)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileUtilNotLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	a, _ := NewAppender(benchLogFile, "rotate", false, "flush", 4096)
	fs := l4g.NewFilters().Add("file", l4g.INFO, a)
	sl.SetFilters(fs)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	sl.SetFilters(nil)
	fs.Close()
	os.Remove(benchLogFile)
}

// Benchmark results (darwin amd64 6g)
// BenchmarkFileLog-4                        300000              7600 ns/op
// BenchmarkFileNotLogged-4                20000000               117 ns/op
// BenchmarkFileUtilLog-4                    300000              7759 ns/op
// BenchmarkFileUtilNotLog-4               10000000               121 ns/op
// BenchmarkCacheFileLog-4                  1000000              1865 ns/op
// BenchmarkCacheFileNotLogged-4           20000000               118 ns/op
// BenchmarkCacheFileUtilLog-4              1000000              1791 ns/op
// BenchmarkCacheFileUtilNotLog-4          10000000               120 ns/op
