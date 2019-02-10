// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bufio"
	"os"
	"sync"
)

var (
	// FileFlushDefault is the default flush size.
	FileFlushDefault = os.Getpagesize() * 2
	// FilePermDefault is the default log file and directory permission.
	// permission to:  owner      group      other
	//                 /```\      /```\      /```\
	// octal:            6          6          6
	// binary:         1 1 0      1 1 0      1 1 0
	// what to permit: r w x      r w x      r w x
	// binary         - 1: enabled, 0: disabled
	// what to permit - r: read, w: write, x: execute
	// permission to  - owner: the user that create the file/folder
	//                  group: the users from group that owner is member
	//                  other: all other users
	FilePermDefault = os.FileMode(0660)
	// FileLineSize is the average line size when calculating lines by file size
	FileLineSize = 256
)

// FileBufWriter represents the buffered writer with lock, current size and lines.
type FileBufWriter struct {
	sync.RWMutex
	*bufio.Writer
	file     *os.File
	name     string
	flush    int
	cursize  int
	curlines int
}

// NewFileBufWriter creates an active buffered writer.
// The file should opened on demand.
func NewFileBufWriter(fname string) *FileBufWriter {
	return &FileBufWriter{
		name:  fname,
		flush: FileFlushDefault,
	}
}

// Close active buffered writer.
func (fbw *FileBufWriter) Close() error {
	fbw.Flush()

	fbw.Lock()
	defer func() {
		fbw.cursize = 0
		fbw.curlines = 0
		fbw.file = nil
		fbw.Writer = nil
		fbw.Unlock()
	}()

	if fbw.file != nil {
		fbw.file.Close()
	}
	return nil
}

// Write bytes to file and calculate current size and lines.
func (fbw *FileBufWriter) Write(b []byte) (n int, err error) {
	fbw.Lock()
	defer fbw.Unlock()

	if fbw.file == nil {
		file, err := os.OpenFile(fbw.name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, FilePermDefault)
		if err != nil {
			LogLogError(err)
			return 0, err
		}

		fbw.file = file
		fbw.cursize = 0
		fbw.curlines = 0
		fbw.cursize = fbw.Size()
		fbw.curlines = fbw.Lines()

		if fbw.flush > 0 {
			fbw.Writer = bufio.NewWriterSize(fbw.file, fbw.flush)
		}
	}

	if fbw.Writer != nil {
		n, err = fbw.Writer.Write(b)
	} else {
		n, err = fbw.file.Write(b)
	}
	if err == nil {
		fbw.cursize += n
		fbw.curlines++
	}
	return n, err
}

// Flush to file.
func (fbw *FileBufWriter) Flush() {
	fbw.Lock()
	defer fbw.Unlock()

	if fbw.Writer != nil {
		fbw.Writer.Flush()
		return
	}
	if fbw.file != nil {
		fbw.file.Sync()
	}
}

// Size returns the size of current file.
func (fbw *FileBufWriter) Size() int {
	if fbw.cursize <= 0 {
		fi, err := fbw.Stat()
		if err == nil {
			fbw.cursize = int(fi.Size())
		}
	}
	return fbw.cursize
}

// Lines returns the lines of current file.
func (fbw *FileBufWriter) Lines() int {
	if fbw.curlines <= 0 {
		fbw.curlines = fbw.Size() / FileLineSize
	}
	return fbw.curlines
}

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (fbw *FileBufWriter) Stat() (os.FileInfo, error) {
	if fbw.file != nil {
		return fbw.file.Stat()
	}
	return os.Stat(fbw.name)
}

// Name returns the name of the file as presented to Open.
func (fbw *FileBufWriter) Name() string {
	return fbw.name
}

// SetFlush sets the flush size.
// If flush <= 0, no buffer.
func (fbw *FileBufWriter) SetFlush(flush int) *FileBufWriter {
	fbw.Close()
	fbw.flush = flush
	return fbw
}
