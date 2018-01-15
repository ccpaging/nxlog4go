// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package colorlog

import (
	"io"
	"os"
	"bytes"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

var HasColor = (os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb") ||
	 os.Getenv("ConEmuANSI") == "ON"

// 0, Black; 1, Red; 2, Green; 3, Yellow; 4, Blue; 5, Purple; 6, Cyan; 7, White

var ColorBytes = [...][]byte{
	[]byte("\x1b[0;34m"),	   // FINEST, Blue
	[]byte("\x1b[0;36m"),	   // FINE, Cyan
	[]byte("\x1b[0;32m"),	   // DEBUG, Green
	[]byte("\x1b[0;35m"), 	   // TRACE, Purple
 	[]byte{},				   // INFO, Default
 	[]byte("\x1b[1;33m"), 	   // WARNING, Yellow
 	[]byte("\x1b[0;31m"), 	   // ERROR, Red
 	[]byte("\x1b[0;31m;47m"),  // CRITICAL, Red - White
}
var ColorReset = []byte("\x1b[0m")

// This is the writer with ANSI color that prints to stderr.
// Support ANSI term only. Use ConEmu in windows
type ColorLogWriter struct {
	mu		sync.Mutex // ensures atomic writes; protects the following fields
	out		io.Writer  // destination for output
	formatSlice [][]byte // Split the format into pieces by % signs
}

// This creates a new ColorLogWriter.
func NewLogWriter() *ColorLogWriter {
	return &ColorLogWriter {
		out:	os.Stderr,
		formatSlice: bytes.Split([]byte(l4g.FORMAT_DEFAULT), []byte{'%'}),
	}
}

// SetOutput sets the output destination for ColorLogWriter.
func (clw *ColorLogWriter) SetOutput(w io.Writer) *ColorLogWriter {
	clw.mu.Lock()
	defer clw.mu.Unlock()
	clw.out = w
	return clw
}

func (clw *ColorLogWriter) Close() {
}

func (clw *ColorLogWriter) LogWrite(rec *l4g.LogRecord) {
	clw.mu.Lock()
	defer clw.mu.Unlock()

	if HasColor {
		clw.out.Write(ColorBytes[rec.Level])
	}
	clw.out.Write(l4g.FormatLogRecord(clw.formatSlice, rec))
	if HasColor {
		clw.out.Write(ColorReset)
	}
}

// Set option. chainable
func (clw *ColorLogWriter) Set(name string, v interface{}) *ColorLogWriter {
	clw.SetOption(name, v)
	return clw
}

// Set option. checkable
func (clw *ColorLogWriter) SetOption(name string, v interface{}) error {
	clw.mu.Lock()
	defer clw.mu.Unlock()

	switch name {
	case "format":
		if format, ok := v.(string); ok {
			clw.formatSlice = bytes.Split([]byte(format), []byte{'%'})
		} else if format, ok := v.([]byte); ok {
				clw.formatSlice = bytes.Split(format, []byte{'%'})
		} else {
			return l4g.ErrBadValue
		}
	default:
		return l4g.ErrBadOption
	}
	return nil
}
