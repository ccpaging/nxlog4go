// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/cast"
)

var (
	dayToSecs int64 = 86400
)

// FileAppender represents the log appender that sends output to a file
type FileAppender struct {
	mu       sync.Mutex         // ensures atomic writes; protects the following fields
	rec      chan *l4g.Recorder // entry channel
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	level  int
	layout l4g.Layout // format entry for output

	out    *RotateFile
	rotate int // rolling number. -1, no rotate; 0, no backup; 1 ... n, backup n log files

	// If cycle < dayToSecs, check interval in seconds.
	// Otherwise, check every (cycle / dayToSecs) day
	cycle int64
	delay int64 // delay seconds since midnight
	reset chan time.Time
}

/* Bytes Buffer */
var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	l4g.Register("file", &FileAppender{})
}

// NewFileAppender creates a new file appender which writes to the file
// named '<exe full path base name>.log'., and without rotating as default.
func NewFileAppender(filename string, args ...interface{}) (*FileAppender, error) {
	if filename == "" {
		base := filepath.Base(os.Args[0])
		filename = strings.TrimSuffix(base, filepath.Ext(base)) + ".log"
	}

	fa := &FileAppender{
		rec: make(chan *l4g.Recorder, 32),

		layout: l4g.NewPatternLayout(""),

		out: NewRotateFile(filename),

		rotate: -1,

		cycle: dayToSecs,
		reset: make(chan time.Time, 8),
	}

	fa.Set(args...)
	return fa, nil
}

// Open creates a new appender which writes to the file
// named '<exe full path base name>.log', and without rotating as default.
func (*FileAppender) Open(filename string, args ...interface{}) (l4g.Appender, error) {
	return NewFileAppender(filename, args...)
}

// Set options.
// Return FileAppender interface.
func (fa *FileAppender) Set(args ...interface{}) l4g.Appender {
	ops, idx, _ := l4g.ArgsToMap(args)
	for _, k := range idx {
		fa.SetOption(k, ops[k])
	}
	return fa
}

// Enabled encodes log Recorder and output it.
func (fa *FileAppender) Enabled(r *l4g.Recorder) bool {
	if r.Level < ca.level {
		return false
	}

	fa.runOnce.Do(func() {
		fa.waitExit = &sync.WaitGroup{}
		fa.waitExit.Add(1)
		go fa.run(fa.waitExit)
	})

	if fa.waitExit == nil {
		// Closed alreay
		fa.output(r)
		return false
	}

	fa.rec <- r
	return false
}

// Write is the filter's output method. This will block if the output
// buffer is full.
func (fa *FileAppender) Write(b []byte) (int, error) {
	return 0, nil
}

func newRotateTimer(cycle int64, delay int64) *time.Timer {
	if cycle <= 0 {
		cycle = dayToSecs
	}

	if cycle < dayToSecs {
		d := time.Duration(cycle) * time.Second
		l4g.LogLogTrace("rotate after %v", d)
		return time.NewTimer(d)
	}

	// now + cycle
	t := time.Now().Add(time.Duration(cycle) * time.Second)
	// back to midnight
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	// midnight + cycle % 86400
	t = t.Add(time.Duration(delay) * time.Second)
	d := t.Sub(time.Now())
	l4g.LogLogTrace("rotate after %v", d)
	return time.NewTimer(d)
}

func (fa *FileAppender) run(waitExit *sync.WaitGroup) {
	l4g.LogLogTrace("cycle %v", time.Duration(fa.cycle)*time.Second)
	fa.doRotate()

	t := newRotateTimer(fa.cycle, fa.delay)

	for {
		select {
		case <-t.C:
			t = newRotateTimer(fa.cycle, fa.delay)
			fa.doRotate()

		case r, ok := <-fa.rec:
			if !ok {
				waitExit.Done()
				return
			}
			fa.output(r)

		case _, ok := <-fa.reset:
			if !ok {
				return
			}
			t = newRotateTimer(fa.cycle, fa.delay)
			l4g.LogLogTrace("Reset cycle %d, delay %d", fa.cycle, fa.delay)
		}
	}
}

func (fa *FileAppender) closeChannel() {
	// notify closing. See run()
	close(fa.rec)
	// waiting for running channel closed
	fa.waitExit.Wait()
	fa.waitExit = nil
	// drain channel
	for left := range fa.rec {
		fa.output(left)
	}
}

// Close is nothing to do here.
func (fa *FileAppender) Close() {
	if fa.waitExit == nil {
		return
	}
	fa.closeChannel()

	fa.mu.Lock()
	defer fa.mu.Unlock()

	close(fa.reset)
	fa.out.Close()
}

func (fa *FileAppender) output(r *l4g.Recorder) {
	if r == nil {
		return
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	fa.mu.Lock()
	fa.layout.Encode(buf, r)
	fa.out.Write(buf.Bytes())
	fa.mu.Unlock()

	if fa.cycle <= 0 {
		// rotating on demand
		fa.doRotate()
	}
}

func (fa *FileAppender) doRotate() {
	if fa.rotate < 0 {
		return
	}

	fa.mu.Lock()
	defer fa.mu.Unlock()

	fa.out.Rotate(fa.rotate)
}

func (fa *FileAppender) setFileOption(k string, v interface{}) (err error) {
	var (
		s   string
		i64 int64
	)
	switch k {
	case "filename": // DEPRECATED. See Open function's dsn argument
		if s, err = cast.ToString(v); err == nil && len(s) >= 0 {
			fa.out = NewRotateFile(s)
		}
	case "head":
		if s, err = cast.ToString(v); err == nil {
			fa.out.Header = s
		}
	case "foot":
		if s, err = cast.ToString(v); err == nil {
			fa.out.Footer = s
		}
	case "maxsize":
		if i64, err = cast.ToInt64(v); err == nil {
			fa.out.Maxsize = i64
		}
	case "maxlines", "maxrecords":
		if i64, err := cast.ToInt64(v); err == nil {
			fa.out.Maxlines = i64
		}
	default:
		return fmt.Errorf("unknown option name %s, value %#v of type %T", k, v, v)
	}
	return
}

func (fa *FileAppender) setCycleOption(k string, v interface{}) (err error) {
	var i64 int64

	switch k {
	case "cycle":
		if i64, err = cast.ToSeconds(v); err == nil {
			fa.cycle = i64
		}
	case "delay", "clock":
		if i64, err = cast.ToSeconds(v); err == nil && i64 >= 0 {
			fa.delay = i64
		}
	default:
		return fmt.Errorf("unknown option name %s, value %#v of type %T", k, v, v)
	}
	if err == nil {
		// send notify to reset channel
		if fa.waitExit != nil {
			fa.reset <- time.Now()
		}
	}
	return
}

// SetOption sets option with:
//  level    - The output level
//  head     - The header of log file. May includes %D (date) and %T (time).
//  foot     - The trailer of log file.
//  maxsize  - Rotating while size > maxsize
//  maxlines - Rotating while lines > maxsize
//  rotate   - -1, no rotating;
//			   0, rotating but no backup;
//			   n, backup n files
//  cycle    - Rotating cycle inn seconds.
//			   <= 0s, checking and rotating on demand;
//			   < 86400s, checking and rotating at every cycle;
//			   >= 86400, checking and rotating at <delay> seconds since midnight.
//  delay    - Rotating seconds since midnight
//
// Pattern layout options:
//	format	 - Layout format string
//  ...
//
// Return error
func (fa *FileAppender) SetOption(k string, v interface{}) (err error) {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	var (
		n  int
		ok bool
	)

	switch k {
	case "level":
		if n, err = l4g.Level(l4g.INFO).IntE(v); err == nil {
			fa.level = n
		}
	case "filename", "head", "foot", "maxsize", "maxlines", "maxrecords":
		err = fa.setFileOption(k, v)
	case "rotate", "maxbackup":
		if n, err = cast.ToInt(v); err == nil {
			fa.rotate = n
		} else if ok, err = cast.ToBool(v); err == nil {
			if ok {
				fa.rotate = 1
			} else {
				fa.rotate = -1
			}
		}
	case "daily": // DEPRECATED. Replaced with cycle and delay
		if ok, err = cast.ToBool(v); err == nil {
			if ok {
				fa.cycle = dayToSecs
				fa.delay = 0
			} else {
				fa.cycle = 5
				fa.delay = 0
			}
		}
	case "cycle", "delay", "clock":
		err = fa.setCycleOption(k, v)
	default:
		return fa.layout.SetOption(k, v)
	}
	return
}
