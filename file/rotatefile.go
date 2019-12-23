// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/rolling"
)

var (
	// EvaluateLineLength is the average line length when calculating lines by file size
	EvaluateLineLength int64 = 256
	// DefaultRotateSize is the default log file rotate size
	DefaultRotateSize int64 = 1024 * 1024
)

// RotateFile represents the buffered writer with lock, header, footer and rotating.
type RotateFile struct {
	file *rolling.Writer // The opened file buffer writer

	Header string // File header
	Footer string // File trailer

	Maxsize  int64 // Rotate at size
	Maxlines int64 // Rotate at lines

	lines int64 // current lines
}

// NewRotateFile creates a new file writer which writes to the given file and
// has rotation enabled if maxsize > 0.
//
// If size > maxsize, any time a new log file is opened, the old one is renamed
// with a .0.log extension to preserve it.  The various Set key-value pairs can be used
// to configure log rotation based on lines, size.
func NewRotateFile(filename string) *RotateFile {
	rf := &RotateFile{
		file:    rolling.NewWriter(filename, 0),
		Maxsize: DefaultRotateSize,
	}
	rf.lines = rf.file.Size() / EvaluateLineLength
	return rf
}

// Close active RotateFile.
func (rf *RotateFile) Close() {
	if rf.file != nil {
		rf.file.Close()
		rf.file = nil
	}
}

func (rf *RotateFile) writeHeadFoot(format string) {
	buf := bytes.NewBuffer(make([]byte, 0, 64))

	layout := l4g.NewPatternLayout(format)
	layout.Encode(buf, &driver.Recorder{Created: time.Now()})
	rf.file.Write(buf.Bytes())
}

// Write binaries to the file.
// It will rotate files if necessary
func (rf *RotateFile) Write(bb []byte) (n int, err error) {
	if len(rf.Header) > 0 && rf.file.Size() == 0 {
		rf.writeHeadFoot(rf.Header)
	}

	n, err = rf.file.Write(bb)
	if err == nil {
		rf.lines++
	}
	return
}

func backupFiles(name string, temp string, backup int) {
	// May compress new log file here

	l4g.LogLogTrace("Backup %s", temp)

	ext := path.Ext(name)                // like ".log"
	path := name[0 : len(name)-len(ext)] // include dir

	// May create backup directory here

	var (
		n    int
		err  error
		slot string
	)
	for n = 0; n < backup; n++ {
		slot = path + fmt.Sprintf(".%d", n) + ext
		_, err = os.Stat(slot)
		if err != nil {
			break
		}
	}
	if err == nil {
		// Full. Remove last
		os.Remove(slot)
		n--
	}

	// May compress previous log file here

	for ; n > 0; n-- {
		prev := path + fmt.Sprintf(".%d", n-1) + ext
		l4g.LogLogTrace("Rename %s to %s", prev, slot)
		os.Rename(prev, slot)
		slot = prev
	}

	l4g.LogLogTrace("Rename %s to %s", temp, path+".0"+ext)
	os.Rename(temp, path+".0"+ext)
}

// Rotate current log file if necessary
func (rf *RotateFile) Rotate(backup int) {
	// l4g.LogLogTrace("Size %d, Maxsize %d", rf.file.Size(), rf.Maxsize)
	if rf.Maxsize > 0 && rf.file.Size() >= rf.Maxsize {
		l4g.LogLogTrace("Size > %d", rf.Maxsize)
	} else if rf.Maxlines > 0 && rf.lines >= rf.Maxlines {
		l4g.LogLogTrace("lines %d > %d", rf.lines, rf.Maxlines)
	} else {
		return
	}

	// Write footer
	if len(rf.Footer) > 0 && rf.file.Size() > 0 {
		rf.writeHeadFoot(rf.Footer)
	}

	name := rf.file.Name

	l4g.LogLogTrace("Close %s", name)
	rf.file.Close()

	if backup <= 0 {
		os.Remove(name)
		return
	}

	// File existed. File size > maxsize. Rotate
	temp := name + time.Now().Format(".20060102-150405")
	l4g.LogLogTrace("Rename %s to %s", name, temp)
	err := os.Rename(name, temp)
	if err != nil {
		l4g.LogLogError(err)
		return
	}
	rf.lines = 0

	backupFiles(name, temp, backup)
}
