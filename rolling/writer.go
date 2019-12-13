// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package rolling

import (
	"bufio"
	"os"
	"path/filepath"
)

var (
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

	// DefaultFileMode is the default log file and directory permission.
	DefaultFileMode = os.FileMode(0660)
)

// Writer represents the buffered writer with rolling up.
type Writer struct {
	Name    string
	Maxsize int64

	file *os.File
	size int64

	bw      *bufio.Writer
	disable bool
}

// NewWriter creates a new file writer which writes to the given file and
// has rotation enabled if maxsize > 0.
//
// If size > maxsize, any time a new log file is opened, the old one is renamed
// with a .0.log extension to preserve it.  The maxsize can be Set.
func NewWriter(filename string, maxsize int64) *Writer {
	return &Writer{
		Name:    filename,
		Maxsize: maxsize,
	}
}

// for benchmark only
func withoutBuf(filename string, maxsize int64) *Writer {
	return &Writer{
		Name:    filename,
		Maxsize: maxsize,
		disable: true,
	}
}

// Flush to file.
func (w *Writer) Flush() {
	if w.bw != nil {
		w.bw.Flush()
		return
	}
	if w.file != nil {
		w.file.Sync()
	}
}

func (w *Writer) open() error {
	file, err := os.OpenFile(w.Name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, DefaultFileMode)
	if err != nil {
		return err
	}

	w.file = file
	if !w.disable {
		w.bw = bufio.NewWriterSize(w.file, 0)
	}

	w.size = 0
	if fi, err := w.file.Stat(); err == nil {
		w.size = fi.Size()
	}
	return nil
}

func (w *Writer) close() (err error) {
	if w.bw != nil {
		w.bw.Flush()
	}

	if w.file != nil {
		err = w.file.Close()
	}

	w.size = 0
	w.file = nil
	w.bw = nil
	return
}

// Close active buffered writer.
func (w *Writer) Close() error {
	return w.close()
}

// Write bytes to file, and rolling up at first if file is ovewlow.
func (w *Writer) Write(b []byte) (n int, err error) {
	if w.Maxsize > 0 && w.size > w.Maxsize {
		w.rolling()
	}

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	if w.bw != nil {
		n, err = w.bw.Write(b)
	} else {
		n, err = w.file.Write(b)
	}

	if err == nil {
		w.size += int64(n)
	}
	return n, err
}

func (w *Writer) rolling() {
	w.close()

	ext := filepath.Ext(w.Name) // like ".log"
	slot := w.Name[0:len(w.Name)-len(ext)] + ".0" + ext
	os.Remove(slot)

	os.Rename(w.Name, slot)
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (w *Writer) Stat() (os.FileInfo, error) {
	if w.file != nil {
		return w.file.Stat()
	}
	return os.Stat(w.Name)
}

// Size return the realtime size of buffered file.
func (w *Writer) Size() int64 {
	if w.file != nil {
		return w.size
	}
	fi, err := os.Stat(w.Name)
	if err != nil {
		return w.size
	}
	return fi.Size()
}
