// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package filelog

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"
	"fmt"

	l4g "github.com/ccpaging/nxlog4go"
)

const testLogFile = "_logtest.log"
const benchLogFile = "_benchlog.log"

var now time.Time = time.Unix(0, 1234567890123456789).In(time.UTC)

func newLogRecord(lvl l4g.Level, src string, msg string) *l4g.LogRecord {
	return &l4g.LogRecord{
		Level:   lvl,
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
		t.Errorf("malformed filelog: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestNextTime(t *testing.T) {
	fmt.Println("now:", time.Now())
	fmt.Println("After 10 minutes:", nextTime(600, -1).Sub(time.Now()))
	fmt.Println("Correct invalid value. cycle = 300 / clock = 0. After 5 minutes:", nextTime(300, 0).Sub(time.Now()))
	fmt.Println("Next midnight:", nextTime(86400, 0).Sub(time.Now()))
	fmt.Println("Next 3:00am:", nextTime(86400, 10800).Sub(time.Now()))
	fmt.Println("Next weekly midnight:", nextTime(86400 * 7, 0).Sub(time.Now()))
}

func BenchmarkFileLog(b *testing.B) {
	sl := l4g.New(l4g.INFO).SetOutput(nil).Set("caller", false)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil).Set("caller", false)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil).Set("caller", false)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil).Set("caller", false)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil).Set("caller", false)
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
	sl := l4g.New(l4g.INFO).SetOutput(nil).Set("caller", false)
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