// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

var loglog *Logger = nil

// Return internal logger.
// This logger used to output log statements from within the package.
// Do not set any filters.
func GetLogLog() *Logger {
	if loglog == nil {
		loglog = New(DEBUG).SetPrefix("nxlog4go").SetPattern("%T %P:%S %L %M\n").SetCaller(false)
	}
	return loglog
}

func LogLogDebug(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(DEBUG) {
		return
	}
	loglog.intLogS(DEBUG, source, intMsg(arg0, args ...))
}

func LogLogTrace(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(TRACE) {
		return
	}
	loglog.intLogS(TRACE, source, intMsg(arg0, args ...))
}

func LogLogInfo(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(INFO) {
		return
	}
	loglog.intLogS(INFO, source, intMsg(arg0, args ...))
}

func LogLogWarn(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(WARNING) {
		return
	}
	loglog.intLogS(WARNING, source, intMsg(arg0, args ...))
}

func LogLogError(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	if loglog.skip(ERROR) {
		return
	}
	loglog.intLogS(ERROR, source, intMsg(arg0, args ...))
}
