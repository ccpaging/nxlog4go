package nxlog4go

import (
	"sync"
	"time"
	"os"
	"fmt"
	"bytes"
	"path"
)

// Rename history log files to "<name>.???.<ext>"
func Backup(newName string, name string, backup int) {
	// May compress new log file here

	ext := path.Ext(name) // like ".log"
	path := name[0:len(name)-len(ext)] // include dir

	// May create backup directory here

	var (
		n int
		err error 
		slot string
	)
	for n = 1; n <= backup; n++ {
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
	
	for ; n > 1; n-- {
		prev := path + fmt.Sprintf(".%d", n - 1) + ext
		os.Rename(prev, slot)
		slot = prev
	}
	
	os.Rename(newName, path + ".1" + ext)
}

// Rotate File buffer writer
type RotateFileWriter struct {
	sync.RWMutex
	// The opened file buffer writer
	*FileBufWriter
	// File header/trailer
	header, footer string
	// Rotate at size
	maxsize   int
	cursize   int
	// Rotate daily
	daily     bool
	currtime  time.Time
	// Keep old files (.1, .2, etc)
	maxbackup int
}

// NewRotateFileWriter creates rotator which writes to the file buffer writer
func NewRotateFileWriter(path string) *RotateFileWriter {
	frw := &RotateFileWriter {
		FileBufWriter: NewFileBufWriter(path),
		daily: false,
	}
	fi, err := frw.FileBufWriter.Stat()
	if err != nil {
		frw.cursize = 0
		frw.currtime = time.Now()
	} else {
		frw.cursize = int(fi.Size())
		frw.currtime = fi.ModTime()
	}
	return frw
}

// Write binaries to the file.
// It will rotate files if necessary
func (frw *RotateFileWriter) Write(bb []byte) (n int, err error) {
	frw.Lock()
	defer frw.Unlock()

	now := time.Now()
	if ((frw.maxsize > 0 && frw.cursize >= frw.maxsize) ||
		(frw.daily && now.Day() != frw.currtime.Day())) {

		frw.Rotate()
	}

	// Write header
	if len(frw.header) > 0 && frw.cursize == 0 {
		fmtSlice := bytes.Split([]byte(frw.header), []byte{'%'})
		n, _ = frw.FileBufWriter.Write(FormatLogRecord(fmtSlice, &LogRecord{Created: time.Now()}))
		frw.cursize += n
	}

	n, err = frw.FileBufWriter.Write(bb)
	frw.cursize += n
	return n, err
}

func (frw *RotateFileWriter) Rotate() {
	defer func() {
		frw.cursize = 0
		frw.currtime = time.Now()
	}()

	// Write footer
	if len(frw.footer) > 0 && frw.cursize > 0 {
		fmtSlice := bytes.Split([]byte(frw.footer), []byte{'%'})
		frw.FileBufWriter.Write(FormatLogRecord(fmtSlice, &LogRecord{Created: time.Now()}))
	}
	// fmt.Fprintf(os.Stderr, "RotateFileWriter(%q): Close file\n", frw.FileBufWriter.Name())
	frw.FileBufWriter.Close() 
	
	name := frw.FileBufWriter.Name()
	if frw.maxbackup <= 0 {
		os.Remove(name)
		return
	}

	// File existed. File size > maxsize. Rotate
	newName := name + time.Now().Format(".20060102-150405")
	err := os.Rename(name, newName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "RotateFileWriter(%q): Rename to %s. %v\n", name, newName, err)
		return
	}
	
	Backup(newName, name, frw.maxbackup)
}

// Set the file header(chainable).  Must be called before the first log
// message is written.  These are formatted similar to the FormatLogRecord (e.g.
// you can use %D and %T in your header for date and time).
func (frw *RotateFileWriter) SetHead(header string) *RotateFileWriter {
	frw.Lock()
	defer frw.Unlock()

	frw.header = header
	return frw
}

// Set the file footer (chainable).  Must be called before the first log
// message is written.  These are formatted similar to the FormatLogRecord (e.g.
// you can use %D and %T in your footer for date and time).
func (frw *RotateFileWriter) SetFoot(footer string) *RotateFileWriter {
	frw.Lock()
	defer frw.Unlock()

	frw.footer = footer
	return frw
}

// Set rotate at size (chainable). Must be called before the first log message
// is written.
func (frw *RotateFileWriter) SetMaxSize(maxsize int) *RotateFileWriter {
	frw.Lock()
	defer frw.Unlock()

	frw.maxsize = maxsize
	return frw
}

// Set rotate daily (chainable). Must be called before the first log message is
// written.
func (frw *RotateFileWriter) SetDaily(daily bool) *RotateFileWriter {
	frw.Lock()
	defer frw.Unlock()

	frw.daily = daily
	return frw
}

// Set max backup files. Must be called before the first log message
// is written.
func (frw *RotateFileWriter) SetMaxBackup(maxbackup int) *RotateFileWriter {
	frw.Lock()
	defer frw.Unlock()

	frw.maxbackup = maxbackup
	return frw
}
