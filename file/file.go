// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package file

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
	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

var (
	dayToSecs int64 = 86400
)

// Appender represents the log appender that sends output to a file
type Appender struct {
	mu       sync.Mutex            // ensures atomic writes; protects the following fields
	rec      chan *driver.Recorder // entry channel
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	level  int
	layout driver.Layout // format entry for output

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

	driver.Register("file", &Appender{})
}

// NewAppender creates a new file appender which writes to the file
// named '<exe full path base name>.log'., and without rotating as default.
func NewAppender(filename string, args ...interface{}) (*Appender, error) {
	if filename == "" {
		base := filepath.Base(os.Args[0])
		filename = strings.TrimSuffix(base, filepath.Ext(base)) + ".log"
	}

	fa := &Appender{
		rec: make(chan *driver.Recorder, 32),

		layout: patt.NewLayout(""),

		out: NewRotateFile(filename),

		rotate: -1,

		cycle: dayToSecs,
		reset: make(chan time.Time, 8),
	}

	fa.SetOptions(args...)
	return fa, nil
}

// Open creates a new appender which writes to the file
// named '<exe full path base name>.log', and without rotating as default.
func (*Appender) Open(filename string, args ...interface{}) (driver.Appender, error) {
	return NewAppender(filename, args...)
}

// Layout returns the output layout for the appender.
func (fa *Appender) Layout() driver.Layout {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	return fa.layout
}

// SetLayout sets the output layout for the appender.
func (fa *Appender) SetLayout(layout driver.Layout) *Appender {
	fa.mu.Lock()
	defer fa.mu.Unlock()
	fa.layout = layout
	return fa
}

// SetOptions sets name-value pair options.
//
// Return the appender.
func (fa *Appender) SetOptions(args ...interface{}) *Appender {
	ops, idx, _ := driver.ArgsToMap(args)
	for _, k := range idx {
		fa.Set(k, ops[k])
	}
	return fa
}

// Enabled encodes log Recorder and output it.
func (fa *Appender) Enabled(r *driver.Recorder) bool {
	// r.Level < fa.level
	if fa.level != 0 && r.Level < fa.level {
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
func (fa *Appender) Write(b []byte) (int, error) {
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

func (fa *Appender) run(waitExit *sync.WaitGroup) {
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

func (fa *Appender) closeChannel() {
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
func (fa *Appender) Close() {
	if fa.waitExit == nil {
		return
	}
	fa.closeChannel()

	fa.mu.Lock()
	defer fa.mu.Unlock()

	close(fa.reset)
	fa.out.Close()
}

func (fa *Appender) output(r *driver.Recorder) {
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

func (fa *Appender) doRotate() {
	if fa.rotate < 0 {
		return
	}

	fa.mu.Lock()
	defer fa.mu.Unlock()

	fa.out.Rotate(fa.rotate)
}

func (fa *Appender) setFileOption(k string, v interface{}) (err error) {
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

func (fa *Appender) setCycleOption(k string, v interface{}) (err error) {
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

// Set sets name-value option with:
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
//	pattern	 - Layout format string
//  ...
//
// Return error
func (fa *Appender) Set(k string, v interface{}) (err error) {
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
		return fa.layout.Set(k, v)
	}
	return
}
