// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package console

import (
	"bytes"
	"io"
	"os"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/internal/cast"
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
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	out    io.Writer  // destination for output
	layout l4g.Layout // format record for output
	color  bool
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

// NewAppender creates a new Appender.
func NewAppender(w io.Writer, args ...interface{}) (*Appender, error) {
	a := &Appender{
		out:    w,
		layout: l4g.NewPatternLayout(""),
		color:  false,
	}
	a.Set(args...)
	return a, nil
}

// Open creates a new appender which writes to stderr.
func (*Appender) Open(dsn string, args ...interface{}) (l4g.Appender, error) {
	return NewAppender(os.Stderr, args...)
}

// SetOutput sets the output destination for Appender.
func (a *Appender) SetOutput(w io.Writer) l4g.Appender {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.out = w
	return a
}

// Close is nothing to do here.
func (a *Appender) Close() {
}

// Write a log recorder to stderr.
func (a *Appender) Write(e *l4g.Entry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.color {
		level := e.Level
		if level >= len(ColorBytes) {
			level = l4g.INFO
		}
		a.out.Write(ColorBytes[level])
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	a.layout.Encode(buf, e)
	a.out.Write(buf.Bytes())

	if a.color {
		a.out.Write(ColorReset)
	}
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
//	color    - Force to color or not
//
// Pattern layout options (The default is JSON):
//	pattern	 - Layout format pattern
//  ...
//
// Return error
func (a *Appender) SetOption(k string, v interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch k {
	case "color":
		if color, err := cast.ToBool(v); err == nil {
			a.color = color
		}
	default:
		return a.layout.SetOption(k, v)
	}

	return nil
}
