// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

var loglog *Logger

// GetLogLog creates internal logger if not existed, and returns it.
// This logger used to output log statements from within the package.
// Do not set any filters.
func GetLogLog() *Logger {
	if loglog == nil {
		loglog = New(DEBUG).Set("prefix", "lg4g").Set("pattern", "%T %P %L %M\n").Set("caller", false)
	}
	return loglog
}

// LogLogDebug logs a message at the debug log level.
func LogLogDebug(arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(DEBUG) {
		return
	}
	loglog.intLog(DEBUG, intMsg(arg0, args...))
}

// LogLogTrace logs a message at the trace log level.
func LogLogTrace(arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(TRACE) {
		return
	}
	loglog.intLog(TRACE, intMsg(arg0, args...))
}

// LogLogInfo logs a message at the info log level.
func LogLogInfo(arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(INFO) {
		return
	}
	loglog.intLog(INFO, intMsg(arg0, args...))
}

// LogLogWarn logs a message at the warn log level.
func LogLogWarn(arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(WARNING) {
		return
	}
	loglog.intLog(WARNING, intMsg(arg0, args...))
}

// LogLogError logs a message at the error log level.
func LogLogError(arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(ERROR) {
		return
	}
	loglog.intLog(ERROR, intMsg(arg0, args...))
}
