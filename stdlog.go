// Copyright (c) 2016 Uber Technologies, Inc.
// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"log"
)

// NewStdLog returns a *log.Logger which writes to the supplied logger Entry at
// Info Level.
func NewStdLog(e *Entry) *log.Logger {
	logFunc := e.Info
	return log.New(&logWriter{logFunc}, "" /* prefix */, 0 /* flags */)
}

// NewStdLogAt returns *log.Logger which writes to supplied logger Entry at
// required level.
func NewStdLogAt(e *Entry, level interface{}) *log.Logger {
	logFunc := levelToFunc(e, level)
	return log.New(&logWriter{logFunc}, "" /* prefix */, 0 /* flags */)
}

// RedirectStdLog redirects output from the standard library's package-global
// logger to the supplied logger at InfoLevel. Since zap already handles caller
// annotations, timestamps, etc., it automatically disables the standard
// library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func RedirectStdLog(e *Entry) func() {
	return redirectStdLogAt(e, INFO)
}

// RedirectStdLogAt redirects output from the standard library's package-global
// logger to the supplied logger at the specified level. Since zap already
// handles caller annotations, timestamps, etc., it automatically disables the
// standard library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func RedirectStdLogAt(e *Entry, level interface{}) func() {
	return redirectStdLogAt(e, level)
}

func redirectStdLogAt(e *Entry, level interface{}) func() {
	flags := log.Flags()
	prefix := log.Prefix()
	writer := log.Writer()
	log.SetFlags(0)
	log.SetPrefix("")
	logFunc := levelToFunc(e.AddCallerSkip(3), level)
	log.SetOutput(&logWriter{logFunc})
	return func() {
		log.SetFlags(flags)
		log.SetPrefix(prefix)
		log.SetOutput(writer)
	}
}

func levelToFunc(e *Entry, level interface{}) func(interface{}, ...interface{}) {
	n := Level(INFO).Int(level)
	switch n {
	case FINEST:
		return e.Finest
	case FINE:
		return e.Fine
	case DEBUG:
		return e.Debug
	case TRACE:
		return e.Trace
	case INFO:
		return e.Info
	case WARN:
		return e.Warn
	case ERROR:
		return e.Error
	case CRITICAL:
		return e.Critical
	}
	return e.Info
}

type logWriter struct {
	logFunc func(msg interface{}, args ...interface{})
}

func (l *logWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	l.logFunc(string(p))
	return len(p), nil
}
