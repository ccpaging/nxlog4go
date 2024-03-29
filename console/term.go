// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package console

import (
	"bytes"
	"io"
	"os"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/cast"
	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

// ColorBytes represents ANSI code to set different color of levels
// 0, Black; 1, Red; 2, Green; 3, Yellow; 4, Blue; 5, Purple; 6, Cyan; 7, White
var ColorBytes = [...][]byte{
	[]byte("\033[30;1m"), // FINEST, Gray
	[]byte("\033[32m"),   // FINE, Green
	[]byte("\033[35m"),   // DEBUG, Magenta
	[]byte("\033[36m"),   // TRACE, Cyan
	nil,                  // INFO, Default
	[]byte("\033[33;1m"), // WARN, LightYellow
	[]byte("\033[31m"),   // ERROR, Red
	[]byte("\033[31;1m"), // CRITICAL, LightRed
}

// ColorReset represents ANSI code to reset color
var ColorReset = []byte("\x1b[0m")

// Appender is an Appender with ANSI color that prints to stderr.
// Support ANSI term includes ConEmu for windows.
type Appender struct {
	mu       sync.Mutex            // ensures atomic writes; protects the following fields
	rec      chan *driver.Recorder // entry channel
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	level  int
	layout driver.Layout // format entry for output

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

	driver.Register("console", &Appender{})
}

// NewAppender creates the appender output to os.Stderr.
func NewAppender(w io.Writer, args ...interface{}) *Appender {
	ca := &Appender{
		rec: make(chan *driver.Recorder, 32),

		layout: patt.NewLayout(""),

		out:   os.Stderr,
		color: false,
	}
	ca.SetOptions(args...)
	return ca
}

// Open creates a new appender which writes to stderr.
func (*Appender) Open(dsn string, args ...interface{}) (driver.Appender, error) {
	return NewAppender(os.Stderr, args...), nil
}

// Writer returns the output destination for the appender.
func (ca *Appender) Writer() io.Writer {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return ca.out
}

// SetOutput sets the output destination for the appender.
func (ca *Appender) SetOutput(w io.Writer) *Appender {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.out = w
	return ca
}

// Layout returns the output layout for the appender.
func (ca *Appender) Layout() driver.Layout {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return ca.layout
}

// SetLayout sets the output layout for the appender.
func (ca *Appender) SetLayout(layout driver.Layout) *Appender {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.layout = layout
	return ca
}

// SetOptions sets name-value pair options.
//
// Return the appender.
func (ca *Appender) SetOptions(args ...interface{}) *Appender {
	ops, idx, _ := driver.ArgsToMap(args...)
	for _, k := range idx {
		ca.Set(k, ops[k])
	}
	return ca
}

// Enabled encodes log Recorder and output it.
func (ca *Appender) Enabled(r *driver.Recorder) bool {
	if r.Level < ca.level {
		return false
	}

	ca.runOnce.Do(func() {
		ca.waitExit = &sync.WaitGroup{}
		ca.waitExit.Add(1)
		go ca.run(ca.waitExit)
	})

	// Write after closed
	if ca.waitExit == nil {
		ca.output(r)
		return false
	}

	ca.rec <- r
	return false
}

// Write is the filter's output method. This will block if the output
// buffer is full.
func (ca *Appender) Write(b []byte) (int, error) {
	return 0, nil
}

func (ca *Appender) run(waitExit *sync.WaitGroup) {
	for {
		select {
		case r, ok := <-ca.rec:
			if !ok {
				waitExit.Done()
				return
			}
			ca.output(r)
		}
	}
}

func (ca *Appender) closeChannel() {
	// notify closing. See run()
	close(ca.rec)
	// waiting for running channel closed
	ca.waitExit.Wait()
	ca.waitExit = nil
	// drain channel
	for r := range ca.rec {
		ca.output(r)
	}
}

// Close is nothing to do here.
func (ca *Appender) Close() {
	if ca.waitExit == nil {
		return
	}
	ca.closeChannel()
}

func (ca *Appender) output(r *driver.Recorder) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	if ca.color {
		level := r.Level
		if level >= len(ColorBytes) {
			level = l4g.INFO
		}
		ca.out.Write(ColorBytes[level])
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	ca.layout.Encode(buf, r)
	ca.out.Write(buf.Bytes())

	if ca.color {
		ca.out.Write(ColorReset)
	}
}

// Set sets name-value option with:
//  level    - The output level
//	color    - Force to color or not
//
// Pattern layout options (The default is JSON):
//	pattern	 - Layout format string
//  ...
//
// Return error.
func (ca *Appender) Set(k string, v interface{}) (err error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	switch k {
	case "level":
		var n int
		if n, err = l4g.Level(l4g.INFO).IntE(v); err == nil {
			ca.level = n
		}
	case "color":
		var color bool
		if color, err = cast.ToBool(v); err == nil {
			ca.color = color
		}
	default:
		return ca.layout.Set(k, v)
	}

	return
}
