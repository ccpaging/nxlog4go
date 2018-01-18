// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"sync"
	"bytes"
	"time"
	"strings"
	
	l4g "github.com/ccpaging/nxlog4go"
)

// Get first rotate time
func nextTime(cycle, clock int) time.Time {
	if cycle <= 0 {
		cycle = 86400
	}
	nrt := time.Now()
	if clock < 0 {
		// Now + cycle
		return nrt.Add(time.Duration(cycle) * time.Second)
	}
	
	// next cycle midnight + clock
	nextCycle := nrt.Add(time.Duration(cycle) * time.Second)
	nrt = time.Date(nextCycle.Year(), nextCycle.Month(), nextCycle.Day(), 
					0, 0, 0, 0, nextCycle.Location())
	return nrt.Add(time.Duration(clock) * time.Second)
}

// This log writer sends output to a file
type FileLogWriter struct {
	mu  sync.Mutex // ensures atomic writes; protects the following fields
	formatSlice [][]byte // Split the format into pieces by % signs

	// 2nd cache, formatted message
	messages chan []byte
	// 3nd cache, destination for output with buffered and rotated
	out *l4g.RotateFileWriter
	// Rotate at size
	maxsize int
	// Rotate cycle in seconds
	cycle, clock int
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
	close(flw.loopReset)
	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closed channel
		time.Sleep(100 * time.Millisecond)
		if len(flw.messages) <= 0 {
			break
		}
	}
	if flw.out != nil {
		flw.out.Close()
	}
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

	nrt := nextTime(flw.cycle, flw.clock)
	rotTimer := time.NewTimer(nrt.Sub(time.Now()))
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
		case <-rotTimer.C:
			nrt = nextTime(flw.cycle, flw.clock)
			rotTimer.Reset(nrt.Sub(time.Now()))
			if flw.cycle > 0 && flw.out.Size() > flw.maxsize {
				flw.out.Rotate()
			}
		case <-flw.loopReset:
			nrt = nextTime(flw.cycle, flw.clock)
			rotTimer.Reset(nrt.Sub(time.Now()))
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
	case "cycle":
		switch value := v.(type) {
		case int:
			flw.cycle = value
		case string:
			// Each with optional fraction and a unit suffix, 
			// such as "300ms", "-1.5h" or "2h45m". 
			// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
			dur, _ := time.ParseDuration(value)
			flw.cycle = int(dur/time.Millisecond)
		default:
			return l4g.ErrBadValue
		}
		if flw.cycle <= 0 {
			flw.out.SetMaxSize(flw.maxsize)
		} else {
			flw.out.SetMaxSize(0)
		}
		if flw.loopRunning {
			flw.loopReset <- time.Now()
		}
	case "clock":
		switch value := v.(type) {
		case int:
			flw.clock = value
		case string:
			dur, _ := time.ParseDuration(value)
			flw.clock = int(dur/time.Millisecond)
		default:
			return l4g.ErrBadValue
		}
		if flw.loopRunning {
			flw.loopReset <- time.Now()
		}
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
		flw.maxsize = maxsize
		if flw.cycle <= 0 {
			flw.out.SetMaxSize(flw.maxsize)
		}
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