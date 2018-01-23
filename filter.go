// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
    "fmt"
	"os"
	"time"
	"errors"
	"strconv"
	"strings"
)

// Various error codes.
var (
    ErrBadOption   = errors.New("invalid or unsupported option")
    ErrBadValue    = errors.New("invalid option value")
)

type FilterProp struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type FilterConfig struct {
	Enabled  string        `xml:"enabled,attr"`
	Tag      string        `xml:"tag"`
	Level    string        `xml:"level"`
	Type     string        `xml:"type"`
	Properties []FilterProp `xml:"property"`
}

type LoggerConfig struct {
	FilterConfigs []FilterConfig `xml:"filter"`
}

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

func ConfigLogWriter(lw LogWriter, props []FilterProp) (LogWriter, bool) {
	good := true
	for _, prop := range props {
		err := lw.SetOption(prop.Name, strings.Trim(prop.Value, " \r\n"))
		if err != nil {
			switch err {
			case ErrBadValue:
				fmt.Fprintf(os.Stderr, "ConfigLogWriter: Bad value of \"%s\"\n", prop.Name)
				good = false
			case ErrBadOption:
				fmt.Fprintf(os.Stderr, "ConfigLogWriter: Unknown property \"%s\"\n", prop.Name)
			default:
			}
		}
	}
	return lw, good
}

/****** Filter ******/

// A Filter represents the log level below which no log records are written to
// the associated LogWriter.
type Filter struct {
	Level Level
	LogWriter

	rec 	chan *LogRecord	// write queue
	closing	bool	// true if filter was closed at API level
}

// Create a new filter
func NewFilter(lvl Level, writer LogWriter) *Filter {
	f := &Filter {
		Level:		lvl,
		LogWriter:	writer,

		rec: 		make(chan *LogRecord, LogBufferLength),
		closing: 	false,
	}

	go f.run()
	return f
}

// This is the filter's output method. This will block if the output
// buffer is full. 
func (f *Filter) writeToChan(rec *LogRecord) {
	if f.closing {
		fmt.Fprintf(os.Stderr, "Filter: channel has been closed. Message is [%s]\n", rec.Message)
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

// Close the filter
func (f *Filter) Close() {
	if f.closing {
		return
	}
	// sleep at most one second and let go routine running
	// drain the log channel before closing
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closing channel
		time.Sleep(100 * time.Millisecond)
		if len(f.rec) <= 0 {
			break
		}
	}

	// block write channel
	f.closing = true

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

// Check filter's configuration
func CheckFilterConfig(fc FilterConfig) (bad bool, enabled bool, lvl Level) {
	bad, enabled, lvl = false, false, INFO

	// Check required children
	if len(fc.Enabled) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required attribute %s\n", "enabled")
		bad = true
	} else {
		enabled = fc.Enabled != "false"
	}
	if len(fc.Tag) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required child <%s>\n", "tag")
		bad = true
	}
	if len(fc.Type) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required child <%s>\n", "type")
		bad = true
	}
	if len(fc.Level) == 0 {
		fmt.Fprintf(os.Stderr, "CheckFilterConfig: Required child <%s>\n", "level")
		bad = true
	}

	switch fc.Level {
	case "FINEST":
		lvl = FINEST
	case "FINE":
		lvl = FINE
	case "DEBUG":
		lvl = DEBUG
	case "TRACE":
		lvl = TRACE
	case "INFO":
		lvl = INFO
	case "WARNING":
		lvl = WARNING
	case "ERROR":
		lvl = ERROR
	case "CRITICAL":
		lvl = CRITICAL
	default:
		fmt.Fprintf(os.Stderr, 
			"CheckFilterConfig: Required child level for filter has unknown value. %s\n", 
			fc.Level)
		bad = true
	}
	return bad, enabled, lvl
}


/****** Filters map ******/

type Filters map[string]*Filter

// Make a new filters
func NewFilters() *Filters {
	return &Filters{}
}

// Add a new filter to the filters map which will only log messages at lvl or
// higher.  This function should not be called from multiple goroutines.
// Returns the logger for chaining.
func (fs Filters) Add(name string, lvl Level, writer LogWriter) *Filters {
	if filt, isExist := fs[name]; isExist {
		filt.Close()
		delete(fs, name)
	}
	fs[name] = NewFilter(lvl, writer)
	return &fs
}

// Close and remove all filters in preparation for exiting the program or a
// reconfiguration of logging.  Calling this is not really imperative, unless
// you want to guarantee that all log messages are written.  Close removes
// all filters (and thus all LogWriters) from the logger.
// Returns the logger for chaining.
func (fs Filters) Close() {
	// Close all filters
	for name, filt := range fs {
		filt.Close()
		delete(fs, name)
	}
}
