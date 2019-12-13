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

// Appender represents the log appender that sends output to a file
type Appender struct {
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

	l4g.Register("file", &Appender{})
}

// NewAppender creates a new file appender which writes to the file
// named '<exe full path base name>.log'., and without rotating as default.
func NewAppender(filename string, args ...interface{}) (*Appender, error) {
	if filename == "" {
		base := filepath.Base(os.Args[0])
		filename = strings.TrimSuffix(base, filepath.Ext(base)) + ".log"
	}

	a := &Appender{
		rec: make(chan *l4g.Recorder, 32),

		level:  l4g.INFO,
		layout: l4g.NewPatternLayout(""),

		out: NewRotateFile(filename),

		rotate: -1,

		cycle: dayToSecs,
		reset: make(chan time.Time, 8),
	}

	a.Set(args...)
	return a, nil
}

// Open creates a new appender which writes to the file
// named '<exe full path base name>.log', and without rotating as default.
func (*Appender) Open(filename string, args ...interface{}) (l4g.Appender, error) {
	return NewAppender(filename, args...)
}

// Set options.
// Return Appender interface.
func (a *Appender) Set(args ...interface{}) l4g.Appender {
	ops, idx, _ := l4g.ArgsToMap(args)
	for _, k := range idx {
		a.SetOption(k, ops[k])
	}
	return a
}

// Enabled encodes log Recorder and output it.
func (a *Appender) Enabled(r *l4g.Recorder) bool {
	if r.Level < a.level {
		return false
	}

	a.runOnce.Do(func() {
		a.waitExit = &sync.WaitGroup{}
		a.waitExit.Add(1)
		go a.run(a.waitExit)
	})

	if a.waitExit == nil {
		// Closed alreay
		a.output(r)
		return false
	}

	a.rec <- r
	return false
}

// Write is the filter's output method. This will block if the output
// buffer is full.
func (a *Appender) Write(b []byte) (int, error) {
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

func (a *Appender) run(waitExit *sync.WaitGroup) {
	l4g.LogLogTrace("cycle %v", time.Duration(a.cycle)*time.Second)
	a.doRotate()

	t := newRotateTimer(a.cycle, a.delay)

	for {
		select {
		case <-t.C:
			t = newRotateTimer(a.cycle, a.delay)
			a.doRotate()

		case r, ok := <-a.rec:
			if !ok {
				waitExit.Done()
				return
			}
			a.output(r)

		case _, ok := <-a.reset:
			if !ok {
				return
			}
			t = newRotateTimer(a.cycle, a.delay)
			l4g.LogLogTrace("Reset cycle %d, delay %d", a.cycle, a.delay)
		}
	}
}

func (a *Appender) closeChannel() {
	// notify closing. See run()
	close(a.rec)
	// waiting for running channel closed
	a.waitExit.Wait()
	a.waitExit = nil
	// drain channel
	for left := range a.rec {
		a.output(left)
	}
}

// Close is nothing to do here.
func (a *Appender) Close() {
	if a.waitExit == nil {
		return
	}
	a.closeChannel()

	a.mu.Lock()
	defer a.mu.Unlock()

	close(a.reset)
	a.out.Close()
}

func (a *Appender) output(r *l4g.Recorder) {
	if r == nil {
		return
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	a.mu.Lock()
	a.layout.Encode(buf, r)
	a.out.Write(buf.Bytes())
	a.mu.Unlock()

	if a.cycle <= 0 {
		// rotating on demand
		a.doRotate()
	}
}

func (a *Appender) doRotate() {
	if a.rotate < 0 {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.out.Rotate(a.rotate)
}

func (a *Appender) notifyReset() {
	if a.waitExit != nil {
		a.reset <- time.Now()
	}
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
func (a *Appender) SetOption(k string, v interface{}) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var (
		s   string
		i   int
		i64 int64
		ok  bool
	)

	switch k {
	case "level":
		if _, ok = v.(int); ok {
			a.level = v.(int)
		} else if _, ok = v.(string); ok {
			a.level = l4g.Level(0).Int(v.(string))
		} else {
			err = fmt.Errorf("can not set option name %s, value %#v of type %T", k, v, v)
		}
	case "filename": // DEPRECATED. See Open function's dsn argument
		if s, err = cast.ToString(v); err == nil && len(s) == 0 {
			a.out = NewRotateFile(s)
		}
	case "head":
		if s, err = cast.ToString(v); err == nil {
			a.out.Header = s
		}
	case "foot":
		if s, err = cast.ToString(v); err == nil {
			a.out.Footer = s
		}
	case "maxsize":
		if i64, err = cast.ToInt64(v); err == nil {
			a.out.Maxsize = i64
		}
	case "maxlines", "maxrecords":
		if i64, err := cast.ToInt64(v); err == nil {
			a.out.Maxlines = i64
		}
	case "rotate", "maxbackup":
		if i, err = cast.ToInt(v); err == nil {
			a.rotate = i
		} else if ok, err = cast.ToBool(v); err == nil {
			if ok {
				a.rotate = 1
			} else {
				a.rotate = -1
			}
		}
	case "daily": // DEPRECATED. Replaced with cycle and delay
		if ok, err = cast.ToBool(v); err == nil {
			if ok {
				a.cycle = dayToSecs
				a.delay = 0
			} else {
				a.cycle = 5
				a.delay = 0
			}
		}
	case "cycle":
		if i64, err = cast.ToSeconds(v); err == nil {
			a.cycle = i64
			a.notifyReset()
		}
	case "delay", "clock":
		if i64, err = cast.ToSeconds(v); err == nil && i64 >= 0 {
			a.delay = i64
			a.notifyReset()
		}
	default:
		return a.layout.SetOption(k, v)
	}
	return
}
