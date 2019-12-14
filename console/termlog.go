// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package console

import (
	"bytes"
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

// ConsoleAppender is an ConsoleAppender with ANSI color that prints to stderr.
// Support ANSI term includes ConEmu for windows.
type ConsoleAppender struct {
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

	l4g.Register("console", &ConsoleAppender{})
}

// NewConsoleAppender creates the appender output to os.Stderr.
func NewConsoleAppender(w io.Writer, args ...interface{}) *ConsoleAppender {
	ca := &ConsoleAppender{
		rec: make(chan *l4g.Recorder, 32),

		level:  l4g.INFO,
		layout: l4g.NewPatternLayout(""),

		out:   os.Stderr,
		color: false,
	}
	ca.Set(args...)
	return ca
}

// Open creates a new appender which writes to stderr.
func (*ConsoleAppender) Open(dsn string, args ...interface{}) (l4g.Appender, error) {
	return NewConsoleAppender(os.Stderr, args...), nil
}

// SetOutput sets the output destination for ConsoleAppender.
func (ca *ConsoleAppender) SetOutput(w io.Writer) l4g.Appender {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.out = w
	return ca
}

// Set options.
// Return ConsoleAppender interface.
func (ca *ConsoleAppender) Set(args ...interface{}) l4g.Appender {
	ops, idx, _ := l4g.ArgsToMap(args)
	for _, k := range idx {
		ca.SetOption(k, ops[k])
	}
	return ca
}

// Enabled encodes log Recorder and output it.
func (ca *ConsoleAppender) Enabled(r *l4g.Recorder) bool {
	if r == nil {
		return false
	}

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
func (ca *ConsoleAppender) Write(b []byte) (int, error) {
	return 0, nil
}

func (ca *ConsoleAppender) run(waitExit *sync.WaitGroup) {
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

func (ca *ConsoleAppender) closeChannel() {
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
func (ca *ConsoleAppender) Close() {
	if ca.waitExit == nil {
		return
	}
	ca.closeChannel()
}

func (ca *ConsoleAppender) output(r *l4g.Recorder) {
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

// SetOption sets option with:
//  level    - The output level
//	color    - Force to color or not
//
// Pattern layout options (The default is JSON):
//	format	 - Layout format string
//  ...
//
// Return error
func (ca *ConsoleAppender) SetOption(k string, v interface{}) (err error) {
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
		return ca.layout.SetOption(k, v)
	}

	return
}
