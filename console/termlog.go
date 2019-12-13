// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package console

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/cast"
)

// ColorBytes represents ANSI code to set different color of levels
// 0, Black; 1, Red; 2, Green; 3, Yellow; 4, Blue; 5, Purple; 6, Cyan; 7, White
var ColorBytes = [...][]byte{
	[]byte("\x1b[0;34m"),     // FINEST, Blue
	[]byte("\x1b[0;36m"),     // FINE, Cyan
	[]byte("\x1b[0;32m"),     // DEBUG, Green
	[]byte("\x1b[0;35m"),     // TRACE, Purple
	nil,                      // INFO, Default
	[]byte("\x1b[1;33m"),     // WARN, Yellow
	[]byte("\x1b[0;31m"),     // ERROR, Red
	[]byte("\x1b[0;31m;47m"), // CRITICAL, Red - White
}

// ColorReset represents ANSI code to reset color
var ColorReset = []byte("\x1b[0m")

// Appender is an Appender with ANSI color that prints to stderr.
// Support ANSI term includes ConEmu for windows.
type Appender struct {
	mu       sync.Mutex         // ensures atomic writes; protects the following fields
	rec      chan *l4g.Recorder // entry channel
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	level  int
	layout l4g.Layout // format entry for output

	out   io.Writer // destination for output
	color bool
}

/* Bytes Buffer */
var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	l4g.Register("console", &Appender{})
}

// NewAppender creates the appender output to os.Stderr.
func NewAppender(w io.Writer, args ...interface{}) *Appender {
	a := &Appender{
		rec: make(chan *l4g.Recorder, 32),

		level:  l4g.INFO,
		layout: l4g.NewPatternLayout(""),

		out:   os.Stderr,
		color: false,
	}
	a.Set(args...)
	return a
}

// Open creates a new appender which writes to stderr.
func (*Appender) Open(dsn string, args ...interface{}) (l4g.Appender, error) {
	return NewAppender(os.Stderr, args...), nil
}

// SetOutput sets the output destination for Appender.
func (a *Appender) SetOutput(w io.Writer) l4g.Appender {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.out = w
	return a
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
	if r == nil {
		return false
	}

	if r.Level < a.level {
		return false
	}

	a.runOnce.Do(func() {
		a.waitExit = &sync.WaitGroup{}
		a.waitExit.Add(1)
		go a.run(a.waitExit)
	})

	// Write after closed
	if a.waitExit == nil {
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

func (a *Appender) run(waitExit *sync.WaitGroup) {
	for {
		select {
		case r, ok := <-a.rec:
			if !ok {
				waitExit.Done()
				return
			}
			a.output(r)
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
	for r := range a.rec {
		a.output(r)
	}
}

// Close is nothing to do here.
func (a *Appender) Close() {
	if a.waitExit == nil {
		return
	}
	a.closeChannel()
}

func (a *Appender) output(r *l4g.Recorder) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.color {
		level := r.Level
		if level >= len(ColorBytes) {
			level = l4g.INFO
		}
		a.out.Write(ColorBytes[level])
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	a.layout.Encode(buf, r)
	a.out.Write(buf.Bytes())

	if a.color {
		a.out.Write(ColorReset)
	}
}

// SetOption sets option with:
//  level    - The output level
//	color    - Force to color or not
//
// Pattern layout options (The default is JSON):
//	format	 - Layout format string
//  ...
//
// Return error
func (a *Appender) SetOption(k string, v interface{}) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch k {
	case "level":
		if _, ok := v.(int); ok {
			a.level = v.(int)
		} else if _, ok := v.(string); ok {
			a.level = l4g.Level(0).Int(v.(string))
		} else {
			err = fmt.Errorf("can not set option name %s, value %#v of type %T", k, v, v)
		}
	case "color":
		var color bool
		if color, err = cast.ToBool(v); err == nil {
			a.color = color
		}
	default:
		return a.layout.SetOption(k, v)
	}

	return
}
