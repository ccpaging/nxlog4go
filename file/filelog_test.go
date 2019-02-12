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

func newLogRecord(level l4g.Level, src string, msg string) *l4g.LogRecord {
	return &l4g.LogRecord{
		Level:   level,
		Source:  src,
		Created: now,
		Message: msg,
	}
}

func TestFileAppender(t *testing.T) {
	w := NewFileAppender(testLogFile, false)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.Init()
	w.Write(newLogRecord(l4g.CRITICAL, "source", "message"))
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
	filters := l4g.NewFilters().Add("file", l4g.FINE, NewFileAppender(testLogFile, false))
	log.SetFilters(filters)
	writeSomethingToLogFile(log)
	log.SetFilters(nil)
	filters.Close()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 158 {
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
	defer l4g.GetLogLog().Set("level", l4g.SILENT)

	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	/* Can also specify manually via the following: (these are the defaults) */
	filter := NewFileAppender(testLogFile, true).Set("maxbackup", 10)
	filter.Set("format", "[%D %T] [%L] (%x) %M%R")
	filter.Set("cycle", 5).Set("clock", -1).Set("maxsize", "5k")

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
	defer l4g.GetLogLog().Set("level", l4g.SILENT)

	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINE).SetOutput(nil)

	/* Can also specify manually via the following: (these are the defaults) */
	filter := NewFileAppender(testLogFile, true).Set("maxbackup", 10)
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

func TestNextTime(t *testing.T) {
	d0, d1 := nextTime(now, 600, -1).Sub(now), time.Duration(10*time.Minute)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (10 minutes): %v should be %v", d0, d1)
	}
	// Correct invalid value cycle = 300，clock = 0 to clock = -1
	// for cycle < 86400
	d0, d1 = nextTime(now, 300, 0).Sub(now), time.Duration(5*time.Minute)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (5 minutes): %v should be %v", d0, d1)
	}

	t1 := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	d0, d1 = nextTime(now, 86400, 0).Sub(now), t1.Sub(now)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (next midnight): %v should be %v", d0, d1)
	}

	t1 = time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
	d0, d1 = nextTime(now, 86400, 10800).Sub(now), t1.Sub(now)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (next 3:00am): %v should be %v", d0, d1)
	}

	t1 = time.Date(now.Year(), now.Month(), now.Day()+7, 0, 0, 0, 0, now.Location())
	d0, d1 = nextTime(now, 86400*7, 0).Sub(now), t1.Sub(now)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (next weekly midnight): %v should be %v", d0, d1)
	}
}

func BenchmarkFileLog(b *testing.B) {
	sl := l4g.NewLogger(l4g.INFO).SetOutput(nil).Set("caller", false)
	b.StopTimer()
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 0))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 0))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 0))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 0))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 4096))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 4096))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 4096))
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
	fs := l4g.NewFilters().Add("file", l4g.INFO, NewFileAppender(benchLogFile, false).Set("flush", 4096))
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
