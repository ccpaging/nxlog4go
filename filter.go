// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
    "fmt"
	"os"
	"time"
	"errors"
	"strconv"
)

// Various error codes.
var (
    ErrBadOption   = errors.New("invalid or unsupported option")
    ErrBadValue    = errors.New("invalid option value")
)
   
/****** LogWriter ******/

// This is an interface for anything that should be able to write logs
type LogWriter interface {
	// Set option about the LogWriter. The options should be set as default.
	// Must be set before the first log message is written if changed.
	// You should test more if have to change options while running.
	SetOption(name string, v interface{}) error

	// This will be called to log a LogRecord message.
	LogWrite(rec *LogRecord)

	// This should clean up anything lingering about the LogWriter, as it is called before
	// the LogWriter is removed.  LogWrite should not be called after Close.
	Close()
}

/****** Filter ******/

// A Filter represents the log level below which no log records are written to
// the associated LogWriter.
type Filter struct {
	Level Level
	LogWriter

	rec 	chan *LogRecord	// write queue
	closed 	bool	// true if Socket was closed at API level
}

func NewFilter(lvl Level, writer LogWriter) *Filter {
	f := &Filter {
		Level:		lvl,
		LogWriter:	writer,

		rec: 		make(chan *LogRecord, DefaultBufferLength),
		closed: 	false,
	}

	go f.run()
	return f
}

// This is the filter's output method. This will block if the output
// buffer is full. 
func (f *Filter) writeToChan(rec *LogRecord) {
	if f.closed {
		fmt.Fprintf(os.Stderr, "LogWriter: channel has been closed. Message is [%s]\n", rec.Message)
		return
	}
	f.rec <- rec
}

func (f *Filter) run() {
	for {
		select {
		case rec, ok := <-f.rec:
			if !ok {
				return
			}
			f.LogWrite(rec)
		}
	}
}

func (f *Filter) Close() {
	if f.closed {
		return
	}
	// sleep at most one second and let go routine running
	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closed channel
		time.Sleep(100 * time.Millisecond)
		if len(f.rec) <= 0 {
			break
		}
	}

	// block write channel
	f.closed = true

	defer f.LogWriter.Close()

	// Notify log writer closing
	close(f.rec)

	if len(f.rec) <= 0 {
		return
	}
	// drain the log channel and write driect
	for rec := range f.rec {
		f.LogWrite(rec)
	}
}

// Parse a number with K/M/G suffixes based on thousands (1000) or 2^10 (1024)
func StrToNumSuffix(str string, mult int) int {
	num := 1
	if len(str) > 1 {
		switch str[len(str)-1] {
		case 'G', 'g':
			num *= mult
			fallthrough
		case 'M', 'm':
			num *= mult
			fallthrough
		case 'K', 'k':
			num *= mult
			str = str[0 : len(str)-1]
		}
	}
	parsed, _ := strconv.Atoi(str)
	return parsed * num
}