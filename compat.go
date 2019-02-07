// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"os"
)

// Panic is compatible with `log`.
func Panic(arg0 interface{}, args ...interface{}) {
	msg := FormatMessage(arg0, args...)
	global.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicln is compatible with `log`.
func Panicln(arg0 interface{}, args ...interface{}) {
	msg := FormatMessage(arg0, args...)
	global.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicf is compatible with `log`.
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	global.intLog(CRITICAL, s)
	panic(s)
}

// Fatal is compatible with `log`.
func Fatal(arg0 interface{}, args ...interface{}) {
	global.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalln is compatible with `log`.
func Fatalln(arg0 interface{}, args ...interface{}) {
	global.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalf is compatible with `log`.
func Fatalf(format string, v ...interface{}) {
	global.intLog(ERROR, fmt.Sprintf(format, v...))
	os.Exit(0)
}

// Print is compatible with `log`.
func Print(arg0 interface{}, args ...interface{}) {
	global.intLog(INFO, arg0, args...)
}

// Println is compatible with `log`.
func Println(arg0 interface{}, args ...interface{}) {
	global.intLog(INFO, arg0, args...)
}

// Printf is compatible with `log`.
func Printf(format string, v ...interface{}) {
	global.intLog(INFO, fmt.Sprintf(format, v...))
}
