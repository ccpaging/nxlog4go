package nxlog4go

import (
	"bufio"
	"os"
	"sync"
)

var (
	// Default flush size of cache writing file
	DefaultFileFlush = os.Getpagesize() * 2
	// Default log file and directory permission
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
	DefaultFilePerm = os.FileMode(0660)
)

// File buffer writer
type FileBufWriter struct {
	sync.RWMutex
	*bufio.Writer
	file *os.File
	name   string
	flush  int
}

func NewFileBufWriter(name string) *FileBufWriter {
	return &FileBufWriter {
		name: name,
		flush: DefaultFileFlush,
	}
}

func (fbw *FileBufWriter) Close() error {
	fbw.Flush()

	fbw.Lock()
	defer func() {
		fbw.file = nil
		fbw.Writer = nil
		fbw.Unlock()
	}()

	if fbw.file != nil {
		fbw.file.Close()
	}
	return nil
}

func (fbw *FileBufWriter) Write(b []byte) (n int, err error) {
	fbw.Lock()
	defer fbw.Unlock()

	if fbw.file == nil {
		file, err := os.OpenFile(fbw.name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, DefaultFilePerm)
		if err != nil {
			return 0, err
		}
	
		fbw.file = file
		if fbw.flush > 0 {
			fbw.Writer = bufio.NewWriterSize(fbw.file, fbw.flush)
		}
	}

	if fbw.Writer != nil {
		return fbw.Writer.Write(b)
	}
	return fbw.file.Write(b)
}

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

func (fbw *FileBufWriter) Stat() (os.FileInfo, error) {
	fbw.Lock()
	defer fbw.Unlock()

	if fbw.file != nil {
		return fbw.file.Stat()
	}
	return os.Stat(fbw.name)
}

func (fbw *FileBufWriter) Name() string {
	return fbw.name
}

func (fbw *FileBufWriter) SetName(name string) *FileBufWriter {
	fbw.Close()
	fbw.name = name
	return fbw
}

// flush <= 0, no bufio
func (fbw *FileBufWriter) SetFlush(flush int) *FileBufWriter {
	fbw.Close()
	fbw.flush = flush
	return fbw
}