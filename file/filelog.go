// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	l4g "github.com/ccpaging/nxlog4go"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileAppender represents the log appender that sends output to a file
type FileAppender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout // format record for output
	// 2nd cache, formatted message
	messages chan []byte
	// 3nd cache, destination for output with buffered and rotated
	out *l4g.RotateFileWriter
	// Rotate cycle in seconds
	cycle, clock int64
	// write loop
	loopInitOnce sync.Once
	loopRunning  bool
	loopReset    chan time.Time
}

// Write log record to channel
func (fa *FileAppender) Write(rec *l4g.LogRecord) {
	fa.loopInitOnce.Do(func() {
		fa.loopRunning = true
		ready := make(chan struct{})
		go fa.writeLoop(ready)
		<-ready
	})

	fa.messages <- fa.layout.Format(rec)
}

// Close drops write loop and closes opened file
func (fa *FileAppender) Close() {
	close(fa.messages)

	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closed channel
		time.Sleep(100 * time.Millisecond)
		if len(fa.messages) <= 0 {
			break
		}
	}
	if fa.out != nil {
		fa.out.Close()
	}

	close(fa.loopReset)
}

func init() {
	l4g.AddAppenderNewFunc("file", New)
	l4g.AddAppenderNewFunc("xml", NewXML)
}

// New creates a new file appender which writes to the file
// named '<exe path base name>.log', and without rotation as default.
func New() l4g.Appender {
	base := filepath.Base(os.Args[0])
	return NewFileAppender(strings.TrimSuffix(base, filepath.Ext(base))+".log", false)
}

// NewXML creates a new file appender which XML format.
func NewXML() l4g.Appender {
	base := filepath.Base(os.Args[0])
	appender := NewFileAppender(strings.TrimSuffix(base, filepath.Ext(base))+".log", false)
	appender.SetOption("head", "<log created=\"%D %T\">%R")

	appender.SetOption("pattern",
		`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%s</source>
		<message>%M</message>
	</record>%R`)

	appender.SetOption("foot", "</log>%R")
	return appender
}

// NewFileAppender creates a new appender which writes to the given file and
// has rotation enabled if maxbackup > 0.
func NewFileAppender(fname string, rotate bool) l4g.Appender {
	return &FileAppender{
		layout:      l4g.NewPatternLayout(l4g.PatternDefault),
		messages:    make(chan []byte, l4g.LogBufferLength),
		out:         l4g.NewRotateFileWriter(fname, rotate),
		cycle:       86400,
		clock:       -1,
		loopRunning: false,
		loopReset:   make(chan time.Time, l4g.LogBufferLength),
	}
}

func newRotateTimer(cycle, clock int64) *time.Timer {
	if cycle <= 0 {
		cycle = 86400
	}
	if cycle < 86400 { // Correct invalid clock
		clock = -1
	}

	if clock < 0 {
		// Now + cycle
		d := time.Duration(cycle) * time.Second
		l4g.LogLogTrace("cycle = %d, clock = %d. Rotate after %v", cycle, clock, d)
		return time.NewTimer(d)
	}

	// now + cycle
	t := time.Now().Add(time.Duration(cycle) * time.Second)
	// back to midnight
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	// midnight + clock
	t = t.Add(time.Duration(clock) * time.Second)
	d := t.Sub(time.Now())
	l4g.LogLogTrace("cycle = %d, clock = %d. Rotate after %v", cycle, clock, d)
	return time.NewTimer(d)
}

func (fa *FileAppender) rotate() {
	if fa.cycle <= 0 {
		return
	}
	if fa.out == nil {
		return
	}
	if !fa.out.IsOverSize() {
		return
	}
	fa.out.Rotate()
}

func (fa *FileAppender) writeLoop(ready chan struct{}) {
	defer func() {
		fa.loopRunning = false
	}()

	fa.rotate()

	t := newRotateTimer(fa.cycle, fa.clock)

	close(ready)
	for {
		select {
		case bb, ok := <-fa.messages:
			if !ok {
				// drain the log channel and write directly
				fa.mu.Lock()
				for bb := range fa.messages {
					fa.out.Write(bb)
				}
				fa.mu.Unlock()
				return
			}

			fa.mu.Lock()
			fa.out.Write(bb)
			if len(fa.messages) <= 0 {
				fa.out.Flush()
			}
			fa.mu.Unlock()

		case <-t.C:
			t = newRotateTimer(fa.cycle, fa.clock)
			fa.rotate()

		case <-fa.loopReset:
			l4g.LogLogTrace("Reset. cycle = %d, clock = %d", fa.cycle, fa.clock)
			t = newRotateTimer(fa.cycle, fa.clock)
		}
	}
}

func (fa *FileAppender) setLoop(k string, v interface{}) error {
	isReset := false

	switch k {
	case "cycle":
		if cycle, err := l4g.ToSeconds(v); err == nil {
			fa.cycle = cycle
			fa.out.Set("rotate", (fa.cycle <= 0))
			isReset = true
		} else {
			return l4g.ErrBadOption
		}
	case "clock", "delay0":
		if clock, err := l4g.ToSeconds(v); err == nil {
			fa.clock = clock
			isReset = true
		} else {
			return l4g.ErrBadOption
		}
	case "daily":
		if daily, err := l4g.ToBool(v); err == nil && daily {
			fa.cycle = 86400
			fa.clock = 0
			fa.out.Set("rotate", false)
			isReset = true
		} else {
			return l4g.ErrBadOption
		}
	default:
		return l4g.ErrBadOption
	}

	if isReset && fa.loopRunning {
		fa.loopReset <- time.Now()
	}
	return nil
}

// Name returns file name.
func (fa *FileAppender) Name() string {
	if fa.out != nil {
		return fa.out.Name()
	}
	return ""
}

// Set option.
// Return Appender interface.
func (fa *FileAppender) Set(name string, v interface{}) l4g.Appender {
	fa.SetOption(name, v)
	return fa
}

// SetOption sets option with:
//	filename  - The opened file
//	flush	  - Flush file cache buffer size
//	maxbackup - Max number for log file storage
//	maxsize	  - Rotate at size
//	maxlines  - Rotate at lines if maxlines > 0
//	pattern	  - Layout format pattern
//	utc		  - Log recorder time zone
//	head 	  - File head format layout pattern
//	foot 	  - File foot format layout pattern
//	cycle	  - The cycle seconds of checking rotate size
//	clock	  - The seconds since midnight
//	daily	  - Checking rotate size at every midnight
// Return errors
func (fa *FileAppender) SetOption(k string, v interface{}) (err error) {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	err = nil

	switch k {
	case "filename":
		fname := ""
		fname, err = l4g.ToString(v)
		if err != nil && len(fname) <= 0 {
			err = l4g.ErrBadValue
		} else {
			// Directory exist already, return nil
			err = os.MkdirAll(filepath.Dir(fname), l4g.FilePermDefault)
			if err == nil {
				// Keep other options
				fa.out.SetFileName(fname)
			}
		}
	case "flush", "head", "foot", "maxbackup", "maxsize", "maxlines":
		err = fa.out.SetOption(k, v)
	case "cycle", "clock", "delay0", "daily":
		err = fa.setLoop(k, v)
	default:
		err = fa.layout.SetOption(k, v)
	}
	return
}
