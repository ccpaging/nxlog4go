// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package colorlog

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
	[]byte("\x1b[1;33m"),     // WARNING, Yellow
	[]byte("\x1b[0;31m"),     // ERROR, Red
	[]byte("\x1b[0;31m;47m"), // CRITICAL, Red - White
}

// ColorReset represents ANSI code to reset color
var ColorReset = []byte("\x1b[0m")

// ColorAppender is an Appender with ANSI color that prints to stderr.
// Support ANSI term includes ConEmu for windows.
type ColorAppender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	out    io.Writer  // destination for output
	layout l4g.Layout // format record for output
	color  bool
}

func init() {
	l4g.AddAppenderNewFunc("color", New)
}

// New creates the default ColorAppender output to os.Stderr.
func New() l4g.Appender {
	return NewColorAppender(os.Stderr)
}

// NewColorAppender creates a new ColorAppender.
func NewColorAppender(w io.Writer) l4g.Appender {
	return &ColorAppender{
		out:    w,
		layout: l4g.NewPatternLayout(l4g.PatternDefault),
		color:  false,
	}
}

// SetOutput sets the output destination for ColorAppender.
func (ca *ColorAppender) SetOutput(w io.Writer) l4g.Appender {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.out = w
	return ca
}

// Close is nothing to do here.
func (ca *ColorAppender) Close() {
}

// Write a log recorder to stderr.
func (ca *ColorAppender) Write(e *l4g.Entry) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	if ca.color {
		level := e.Level
		if int(level) >= len(ColorBytes) {
			level = l4g.INFO
		}
		ca.out.Write(ColorBytes[level])
		ca.out.Write(ca.layout.Format(e))
		ca.out.Write(ColorReset)
	} else {
		ca.out.Write(ca.layout.Format(e))
	}
}

// Set option.
// Return Appender interface.
func (ca *ColorAppender) Set(name string, v interface{}) l4g.Appender {
	ca.SetOption(name, v)
	return ca
}

// SetOption sets option with:
//	color    - Force to color text or not
//	pattern	 - Layout format pattern
//	utc 	 - Log recorder time zone
// Return errors
func (ca *ColorAppender) SetOption(k string, v interface{}) (err error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	err = nil

	switch k {
	case "color":
		color := false
		if color, err = l4g.ToBool(v); err == nil {
			ca.color = color
		}
	default:
		return ca.layout.SetOption(k, v)
	}
	return
}
