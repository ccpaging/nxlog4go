// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"sync"
	"time"
	"os"
	"fmt"
	"path"
)

// Rotate File buffer writer
type RotateFileWriter struct {
	sync.RWMutex
	// The opened file buffer writer
	*FileBufWriter
	// File header/trailer
	header, footer string
	// Rotate at size
	maxsize   int
	// Rotate at linecount
	maxlines  int
	// Rotate daily
	daily     bool
	curtime   time.Time
	// Keep old logfiles (.001, .002, etc)
	rotate    bool
	maxbackup int
}

// NewRotateFileWriter creates rotator which writes to the file buffer writer
func NewRotateFileWriter(fname string, rotate bool) *RotateFileWriter {
	rfw := &RotateFileWriter {
		FileBufWriter: NewFileBufWriter(fname),
		daily: false,
		rotate: rotate,
		maxbackup: 9,
	}
	fi, err := rfw.FileBufWriter.Stat()
	if err != nil {
		rfw.curtime = time.Now()
	} else {
		rfw.curtime = fi.ModTime()
	}
	return rfw
}

// Write binaries to the file.
// It will rotate files if necessary
func (rfw *RotateFileWriter) Write(bb []byte) (n int, err error) {
	rfw.Lock()
	defer rfw.Unlock()

	if rfw.rotate {
		if rfw.daily {
			if rfw.curtime.Day() != time.Now().Day() {
				if rfw.Size() > 0 {
					rfw.Rotate()
				}
			}
		} else {
			if rfw.IsOverSize() {
				rfw.Rotate()
			}
		}
	}

	if len(rfw.header) > 0 && rfw.Size() == 0 {
		layout := NewPatternLayout(rfw.header)
		rfw.FileBufWriter.Write(layout.Format(&LogRecord{Created: time.Now()}))
	}

	return rfw.FileBufWriter.Write(bb)
}

// Rename history log files to "<name>.???.<ext>"
func intBackup(newName string, pattern string, backup int) {
	// May compress new log file here

	LogLogTrace("Backup %s", newName)

	ext := path.Ext(pattern) // like ".log"
	path := pattern[0:len(pattern)-len(ext)] // include dir

	// May create backup directory here

	var (
		n int
		err error 
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
		prev := path + fmt.Sprintf(".%d", n - 1) + ext
		LogLogTrace("Rename %s to %s", prev, slot)
		os.Rename(prev, slot)
		slot = prev
	}
	
	LogLogTrace("Rename %s to %s", newName, path + ".0" + ext)
	os.Rename(newName, path + ".0" + ext)
}

func (rfw *RotateFileWriter) Rotate() {
	defer func() {
		rfw.curtime = time.Now()
	}()

	// Write footer
	if len(rfw.footer) > 0 && rfw.Size() > 0 {
		layout := NewPatternLayout(rfw.footer)
		rfw.FileBufWriter.Write(layout.Format(&LogRecord{Created: time.Now()}))
	}

	LogLogTrace("Close %s", rfw.FileBufWriter.Name())
	rfw.FileBufWriter.Close()
	
	name := rfw.FileBufWriter.Name()
	if rfw.maxbackup <= 0 {
		os.Remove(name)
		return
	}

	// File existed. File size > maxsize. Rotate
	newName := name + time.Now().Format(".20060102-150405")
	LogLogTrace("Rename %s to %s", name, newName)
	err := os.Rename(name, newName)
	if err != nil {
		LogLogError(err)
		return
	}
	
	intBackup(newName, name, rfw.maxbackup)
}

// Set option. chainable
func (rfw *RotateFileWriter) Set(k string, v interface{}) *RotateFileWriter {
	rfw.SetOption(k, v)
	return rfw
}

/*
Set option. checkable. Better be set before SetFilters()
Option names include:
	flush	  - Flush file cache buffer size
	maxbackup - Max number for log file storage
	maxsize	  - Rotate at size
	head 	  - File head format layout pattern
	foot 	  - File foot format layout pattern
	daily	  - Checking rotate size at every midnight
	rotate    - 
*/
func (rfw *RotateFileWriter) SetOption(k string, v interface{}) (err error) {
	err = nil

	switch k {
	case "flush":
		flush := 0
		if flush, err = ToInt(v); err == nil {
			rfw.SetFlush(flush)
		}		
	case "head":
		header := ""
		if header, err = ToString(v); err == nil {
			rfw.header = header
		}
	case "foot":
		footer := ""
		if footer, err = ToString(v); err == nil {
			rfw.footer = footer
		}
	case "maxsize":
		maxsize := 0
		if maxsize, err = ToInt(v); err == nil {
			rfw.maxsize = maxsize
		}
	case "maxlines":
		maxlines := 0
		if maxlines, err = ToInt(v); err == nil {
			rfw.maxlines = maxlines
		}
	case "daily":
		daily := false
		if daily, err = ToBool(v); err == nil {
			rfw.daily = daily
		}
	case "rotate":
		rotate := false
		if rotate, err = ToBool(v); err == nil {
			LogLogTrace("Set rotate to %v", rotate)
			rfw.rotate = rotate
		}
	case "maxbackup":
		maxbackup := 0
		if maxbackup, err = ToInt(v); err == nil {
			rfw.maxbackup = maxbackup
		}
	default:
		err = ErrBadOption
	}
	return
}

// Set the file header(chainable).  Must be called before the first log
// message is written.
func (rfw *RotateFileWriter) SetFileName(path string) *RotateFileWriter {
	rfw.Lock()
	defer rfw.Unlock()
	LogLogTrace("Set file name as %s", path)
	rfw.FileBufWriter.Close()
	rfw.FileBufWriter = NewFileBufWriter(path)
	return rfw
}

func (rfw RotateFileWriter) MaxBackup() int {
	return rfw.maxbackup
}

func (rfw RotateFileWriter) MaxSize() int {
	return rfw.maxsize
}

func (rfw RotateFileWriter) MaxLines() int {
	return rfw.maxlines
}

func (rfw RotateFileWriter) IsOverSize() bool {
	if rfw.maxsize > 0 && rfw.Size() >= rfw.maxsize {
		return true
	} else if rfw.maxlines > 0 && rfw.Lines() >= rfw.maxlines {
		return true
	}
	return false
}
