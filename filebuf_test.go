// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func TestFileWriter(t *testing.T) {
	w := NewFileBufWriter(testLogFile).SetFlush(0)

	defer os.Remove(testLogFile)

	layout := NewPatternLayout(testPattern)
	w.Write(layout.Format(newLogRecord(CRITICAL, "prefix", "source", "message")))
	w.Close()

	runtime.Gosched()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 52 {
		t.Errorf("malformed filelog: %q (%d bytes)", string(contents), len(contents))
	}
}
