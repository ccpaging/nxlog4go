// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

var loglog *Logger = nil

// Return internal logger.
// This logger used to output log statements from within the package.
// Do not set any filters.
func GetLogLog() *Logger {
	if loglog == nil {
		loglog = New(_SILENT_).SetPrefix("nxlog4go").SetPattern("%P:%S %L %M\n").SetCaller(false)
	}
	return loglog
}

func loglogTrace(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	loglog.Log(TRACE, source, arg0, args ...)
}

func loglogWarn(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	loglog.Log(WARNING, source, arg0, args ...)
}

func loglogError(source string, arg0 interface{}, args ...interface{}) {
	if loglog == nil {
		return
	}
	loglog.Log(ERROR, source, arg0, args ...)
}
