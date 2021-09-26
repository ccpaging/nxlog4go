// Copyright (c) 2016 Uber Technologies, Inc.
// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	stdlog "log"
)

// NewStdLog returns a *log.Logger which writes to the supplied logger Entry at
// Info Level.
//
// Example: log := NewStdLogAt(e.AddCallerSkip(3), level)
func NewStdLog(e *Entry) *stdlog.Logger {
	logFunc := e.Info
	return stdlog.New(&logWriter{logFunc}, /* out io.Writer */
		"", /* prefix */
		0 /* flags */)
}

// NewStdLogAt returns *log.Logger which writes to supplied logger Entry at
// required level.
//
// Example: log := NewStdLogAt(e.AddCallerSkip(3), "debug")
func NewStdLogAt(e *Entry, level interface{}) *stdlog.Logger {
	logWriteFunc := levelToWriteFunc(e, level)
	return stdlog.New(&logWriter{logWriteFunc}, /* out io.Writer */
		"", /* prefix */
		0 /* flags */)
}

// RedirectStdLog redirects output from the standard library's package-global
// logger to the supplied logger at Info level. Since Entry already handles caller
// annotations, timestamps, etc., it automatically disables the standard
// library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// output io.Writer to os.Stderr.
//
// Example: restoreFunc := RedirectStdLog(elog.AddCallerSkip(3))
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
//
// Example: restoreFunc := RedirectStdLogAt(elog.AddCallerSkip(3), "debug")
func RedirectStdLogAt(e *Entry, level interface{}) func() {
	return redirectStdLogAt(e, level)
}

func redirectStdLogAt(e *Entry, level interface{}) func() {
	flags := stdlog.Flags()
	prefix := stdlog.Prefix()
	writer := stdlog.Writer()
	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
	writeFunc := levelToWriteFunc(e, level)
	stdlog.SetOutput(&logWriter{writeFunc})

	// Return the restore original function
	return func() {
		stdlog.SetFlags(flags)
		stdlog.SetPrefix(prefix)
		stdlog.SetOutput(writer)
	}
}
func levelToWriteFunc(e *Entry, level interface{}) func(interface{}, ...interface{}) {
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
