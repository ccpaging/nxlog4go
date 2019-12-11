// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package rollingfile

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

var testFiles = []string{"_test.log", "_test.0.log"}
var testString = "hello, world"
var testLongString = "Everything is created now (notice that I will be printing to the file)"
var benchLogFile = "_benchlog.log"

func TestWriter(t *testing.T) {
	testFile := testFiles[0]

	w := &Writer{Name: testFile, Maxsize: 5 * 1024}
	defer os.Remove(testFile)

	w.Write([]byte(testString))
	w.Close()

	runtime.Gosched()

	if contents, err := ioutil.ReadFile(testFile); err != nil {
		t.Errorf("read(%q): %s", testFiles, err)
	} else if len(contents) != 12 {
		t.Errorf("malformed file: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestRolling(t *testing.T) {
	w := &Writer{Name: testFiles[0], Maxsize: 5 * 1024}

	for j := 0; j < 15; j++ {
		for i := 0; i < 200/(j+1); i++ {
			w.Write([]byte(testLongString + "\n"))
		}
	}

	w.Close()

	runtime.Gosched()

	for _, testFile := range testFiles {
		if contents, err := ioutil.ReadFile(testFile); err != nil {
			t.Errorf("read(%q): %s", testFile, err)
		} else if len(contents) != 213 && len(contents) != 5183 {
			t.Errorf("malformed file: %q (%d bytes)", string(contents), len(contents))
		}

		os.Remove(testFile)
	}
}

func BenchmarkWithoutBuf(b *testing.B) {
	w := withoutBuf(benchLogFile, 0)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	b.StopTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		w.Write([]byte(testLongString + "\n"))
	}
	b.StopTimer()
}

func BenchmarkBuffered(b *testing.B) {
	w := NewWriter(benchLogFile, 0)
	defer func() {
		w.Close()
		os.Remove(benchLogFile)
	}()
	b.StopTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		w.Write([]byte(testLongString + "\n"))
	}
	b.StopTimer()
}
