// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package colorlog

import (
	"io"
	"os"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

var isColorful = (os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb") ||
	 os.Getenv("ConEmuANSI") == "ON"

// 0, Black; 1, Red; 2, Green; 3, Yellow; 4, Blue; 5, Purple; 6, Cyan; 7, White
var ColorBytes = [...][]byte{
	[]byte("\x1b[0;34m"),	   // FINEST, Blue
	[]byte("\x1b[0;36m"),	   // FINE, Cyan
	[]byte("\x1b[0;32m"),	   // DEBUG, Green
	[]byte("\x1b[0;35m"), 	   // TRACE, Purple
 	nil,					   // INFO, Default
 	[]byte("\x1b[1;33m"), 	   // WARNING, Yellow
 	[]byte("\x1b[0;31m"), 	   // ERROR, Red
 	[]byte("\x1b[0;31m;47m"),  // CRITICAL, Red - White
}
var ColorReset = []byte("\x1b[0m")

// This is the writer with ANSI color that prints to stderr.
// Support ANSI term only includes ConEmu for windows.
type ColorAppender struct {
	mu		sync.Mutex // ensures atomic writes; protects the following fields
	out		io.Writer  // destination for output
	layout  l4g.Layout // format record for output
}

// This creates a new ColorAppender.
func NewAppender() l4g.Appender {
	return &ColorAppender {
		out:	os.Stderr,
		layout: l4g.NewPatternLayout(l4g.PATTERN_DEFAULT),
	}
}

// SetOutput sets the output destination for ColorAppender.
func (ca *ColorAppender) SetOutput(w io.Writer) l4g.Appender {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.out = w
	return ca
}

func (ca *ColorAppender) Close() {
}

func (ca *ColorAppender) Write(rec *l4g.LogRecord) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	if isColorful {
		ca.out.Write(ColorBytes[rec.Level])
		defer ca.out.Write(ColorReset)
	}
	ca.out.Write(ca.layout.Format(rec))
}

// Set option. chainable
func (ca *ColorAppender) Set(name string, v interface{}) l4g.Appender {
	ca.SetOption(name, v)
	return ca
}

// Set option. checkable
func (ca *ColorAppender) SetOption(name string, v interface{}) error {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	switch name {
	case "pattern":
		if pattern, ok := v.(string); ok {
			ca.layout.Set("pattern", pattern)
		} else if pattern, ok := v.([]byte); ok {
			ca.layout.Set("pattern", pattern)
		} else {
			return l4g.ErrBadValue
		}
	default:
		return l4g.ErrBadOption
	}
	return nil
}
