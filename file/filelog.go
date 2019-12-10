// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/cast"
)

// Appender represents the log appender that sends output to a file
type Appender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout // format record for output
	// 2nd cache, formatted message
	messages chan []byte
	// 3nd cache, destination for output with buffered and rotated
	out *l4g.RotateFileWriter
	// write loop
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	// Rotate cycle in seconds
	cycle int64
	clock int64
	reset chan time.Time
}

// Write log record to channel
func (a *Appender) Write(e *l4g.Entry) {
	a.runOnce.Do(func() {
		a.waitExit = &sync.WaitGroup{}
		a.waitExit.Add(1)
		go a.writeLoop(a.waitExit)
	})

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	a.layout.Encode(buf, e)

	if a.waitExit == nil {
		a.out.Write(buf.Bytes())
	}

	a.messages <- buf.Bytes()
}

// Close drops write loop and closes opened file
func (a *Appender) Close() {
	if a.waitExit == nil {
		return
	}

	close(a.messages)
	// Waiting for running channel closed
	a.waitExit.Wait()
	a.waitExit = nil

	close(a.reset)

	if a.out != nil {
		a.out.Close()
	}
}

/* Bytes Buffer */
var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	l4g.Register("file", &Appender{})
}

// NewAppender creates a new appender which writes to the given file and
// has rotation enabled if maxbackup > 0.
func NewAppender(filename string, args ...interface{}) (*Appender, error) {
	if filename == "" {
		base := filepath.Base(os.Args[0])
		filename = strings.TrimSuffix(base, filepath.Ext(base)) + ".log"
	}

	a := &Appender{
		layout:   l4g.NewPatternLayout(""),
		messages: make(chan []byte, l4g.LogBufferLength),
		out:      l4g.NewRotateFileWriter(filename, false),
		cycle:    86400,
		clock:    -1,
		reset:    make(chan time.Time, l4g.LogBufferLength),
	}

	a.Set(args...)
	return a, nil
}

// Open creates a new appender which writes to the file
// named '<exe full path base name>.log', and without rotating as default.
func (*Appender) Open(filename string, args ...interface{}) (l4g.Appender, error) {
	return NewAppender(filename, args...)
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

func (a *Appender) rotate() {
	if a.cycle <= 0 {
		return
	}
	if a.out == nil {
		return
	}
	if !a.out.IsOverSize() {
		return
	}
	a.out.Rotate()
}

func (a *Appender) writeLoop(waitExit *sync.WaitGroup) {
	a.rotate()
	t := newRotateTimer(a.cycle, a.clock)
	for {
		select {
		case bb, ok := <-a.messages:
			if !ok {
				// drain the log channel and write directly
				a.mu.Lock()
				for bb := range a.messages {
					a.out.Write(bb)
				}
				a.mu.Unlock()
				waitExit.Done()
				return
			}

			a.mu.Lock()
			a.out.Write(bb)
			if len(a.messages) <= 0 {
				a.out.Flush()
			}
			a.mu.Unlock()

		case <-t.C:
			t = newRotateTimer(a.cycle, a.clock)
			a.rotate()

		case <-a.reset:
			l4g.LogLogTrace("Reset. cycle = %d, clock = %d", a.cycle, a.clock)
			t = newRotateTimer(a.cycle, a.clock)
		}
	}
}

func (a *Appender) setLoop(k string, v interface{}) error {
	isReset := false

	switch k {
	case "cycle":
		if cycle, err := cast.ToSeconds(v); err == nil {
			a.cycle = cycle
			a.out.Set("rotate", (a.cycle <= 0))
			isReset = true
		} else {
			return err
		}
	case "clock", "delay0":
		if clock, err := cast.ToSeconds(v); err == nil {
			a.clock = clock
			isReset = true
		} else {
			return err
		}
	case "daily":
		if daily, err := cast.ToBool(v); err == nil && daily {
			a.cycle = 86400
			a.clock = 0
			a.out.Set("rotate", false)
			isReset = true
		} else {
			return err
		}
	case "weekly":
	case "monthly":
	default:
		return l4g.ErrBadOption
	}

	if isReset && a.waitExit != nil {
		a.reset <- time.Now()
	}
	return nil
}

// Name returns file name.
func (a *Appender) Name() string {
	if a.out != nil {
		return a.out.Name()
	}
	return ""
}

// Set options.
// Return Appender interface.
func (a *Appender) Set(args ...interface{}) l4g.Appender {
	ops, index, _ := l4g.ArgsToMap(args)
	for _, k := range index {
		a.SetOption(k, ops[k])
	}
	return a
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
func (a *Appender) SetOption(k string, v interface{}) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	err = nil

	switch k {
	case "filename":
		fname := ""
		fname, err = cast.ToString(v)
		if err != nil && len(fname) <= 0 {
			err = l4g.ErrBadValue
		} else {
			// Directory exist already, return nil
			err = os.MkdirAll(filepath.Dir(fname), l4g.FilePermDefault)
			if err == nil {
				// Keep other options
				a.out.SetFileName(fname)
			}
		}
	case "rotate":
	case "flush", "head", "foot", "maxbackup", "maxsize", "maxlines":
		err = a.out.SetOption(k, v)
	case "cycle", "clock", "delay0", "daily", "weekly", "monthly":
		err = a.setLoop(k, v)
	default:
		err = a.layout.SetOption(k, v)
	}
	return
}
