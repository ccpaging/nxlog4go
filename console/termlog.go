// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package console

import (
	"io"
	"os"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
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

func init() {
	l4g.Register("console", &Appender{})
}

// NewAppender creates a new Appender.
func NewAppender(w io.Writer, args ...interface{}) (*Appender, error) {
	a := &Appender{
		out:    w,
		layout: l4g.NewPatternLayout(l4g.PatternDefault),
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
		if int(level) >= len(ColorBytes) {
			level = l4g.INFO
		}
		a.out.Write(ColorBytes[level])
		a.out.Write(a.layout.Format(e))
		a.out.Write(ColorReset)
	} else {
		a.out.Write(a.layout.Format(e))
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
//	color    - Force to color text or not
//	pattern	 - Layout format pattern
//	utc 	 - Log recorder time zone
// Return errors
func (a *Appender) SetOption(k string, v interface{}) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	err = nil

	switch k {
	case "color":
		color := false
		if color, err = l4g.ToBool(v); err == nil {
			a.color = color
		}
	default:
		return a.layout.SetOption(k, v)
	}
	return
}
