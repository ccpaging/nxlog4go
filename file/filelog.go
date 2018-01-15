// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"sync"
	"bytes"
	"os"
	"time"
	"path"
	"strings"
	
	l4g "github.com/ccpaging/nxlog4go"
)

// This log writer sends output to a file
type FileLogWriter struct {
	mu  sync.Mutex // ensures atomic writes; protects the following fields
	formatSlice [][]byte // Split the format into pieces by % signs

	// 2nd cache, formatted message
	messages chan []byte
	// 3nd cache, destination for output with buffered and rotated
	out *l4g.RotateFileWriter
	// Rotate at size
	maxsize int64
	// Rotate cycle in seconds
	cycle, delay0 int64	
	// write loop
	loopRunning bool
	loopReset chan time.Time
}

func (flw *FileLogWriter) LogWrite(rec *l4g.LogRecord) {
	if !flw.loopRunning {
		flw.loopRunning = true
		go flw.writeLoop()
	}
	flw.messages <- l4g.FormatLogRecord(flw.formatSlice, rec)
}

func (flw *FileLogWriter) Close() {
	close(flw.messages)
	// Loop may not running if no message write yet
	if flw.loopRunning {
		// Waiting at most one second and let go routine exit
		for i := 10; i > 0 && flw.loopRunning == false; i-- {
			// Must call Sleep here, otherwise, may panic send on closed channel
			time.Sleep(100 * time.Millisecond)
		}
	}
	if flw.out != nil {
		flw.out.Close()
	}
	close(flw.loopReset)
}

// NewFileLogWriter creates a new LogWriter which writes to the given file and
// has rotation enabled if maxrotate > 0.
func NewFileLogWriter(path string, maxbackup int) *FileLogWriter {
	return &FileLogWriter{
		formatSlice: bytes.Split([]byte(l4g.FORMAT_DEFAULT), []byte{'%'}),	
		messages: make(chan []byte,  l4g.DefaultBufferLength),
		out: l4g.NewRotateFileWriter(path).SetMaxBackup(maxbackup),
		loopRunning: false,
		loopReset: make(chan time.Time, 5),
	}
}

func (flw *FileLogWriter) writeLoop() {
	defer func() {
		flw.loopRunning = false
	}()

	for {
		select {
		case bb, ok := <-flw.messages:
			flw.mu.Lock()
			flw.out.Write(bb)
			if len(flw.messages) <= 0 {
				flw.out.Flush()
			}
			flw.mu.Unlock()
			
			if !ok {
 				// drain the log channel and write directly
				flw.mu.Lock()
				for bb := range flw.messages {
					flw.out.Write(bb)
				}
				flw.mu.Unlock()
				return
			}
		case <-flw.loopReset:
		}
	}
}

// Set option. chainable
func (flw *FileLogWriter) Set(name string, v interface{}) *FileLogWriter {
	flw.SetOption(name, v)
	return flw
}

// Set option. checkable. Must be set before the first log message is written.
func (flw *FileLogWriter) SetOption(name string, v interface{}) error {
	flw.mu.Lock()
	defer flw.mu.Unlock()

	switch name {
	case "filename":
		if filename, ok := v.(string); ok {
			if len(filename) <= 0 {
				return l4g.ErrBadValue
			}
			err := os.MkdirAll(path.Dir(filename), l4g.DefaultFilePerm)
			if err != nil {
				return err
			}
			// flw.out.SetFileName(filename)
		} else {
			return l4g.ErrBadValue
		}
	case "flush":
		var flush int
		switch value := v.(type) {
		case int:
			flush = value
		case string:
			flush = l4g.StrToNumSuffix(strings.Trim(value, " \r\n"), 1024)
		default:
			return l4g.ErrBadValue
		}
		flw.out.SetFlush(flush)
	case "maxbackup":
		var maxbackup int
		switch value := v.(type) {
		case int:
			maxbackup = value
		case string:
			maxbackup = l4g.StrToNumSuffix(strings.Trim(value, " \r\n"), 1)
		default:
			return l4g.ErrBadValue
		}
		flw.out.SetMaxBackup(maxbackup)
	case "maxsize":
		var maxsize int
		switch value := v.(type) {
		case int:
			maxsize = value
		case string:
			maxsize = l4g.StrToNumSuffix(strings.Trim(value, " \r\n"), 1024)
		default:
			return l4g.ErrBadValue
		}
		flw.out.SetMaxSize(maxsize)
	case "format":
		if format, ok := v.(string); ok {
			flw.formatSlice = bytes.Split([]byte(format), []byte{'%'})
		} else if format, ok := v.([]byte); ok {
				flw.formatSlice = bytes.Split(format, []byte{'%'})
		} else {
			return l4g.ErrBadValue
		}
	case "head":
		if header, ok := v.(string); ok {
			flw.out.SetHead(header)
		} else {
			return l4g.ErrBadValue
		}
	case "foot":
		if footer, ok := v.(string); ok {
			flw.out.SetFoot(footer)
		} else {
			return l4g.ErrBadValue
		}
	default:
		return l4g.ErrBadOption
	}
	return nil
}