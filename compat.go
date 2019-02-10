// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"os"
)

// Panic is compatible with `log`.
func (log Logger) Panic(arg0 interface{}, args ...interface{}) {
	msg := FormatMessage(arg0, args...)
	log.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicln is compatible with `log`.
func (log Logger) Panicln(arg0 interface{}, args ...interface{}) {
	msg := FormatMessage(arg0, args...)
	log.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicf is compatible with `log`.
func (log Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	log.intLog(CRITICAL, s)
	panic(s)
}

// Fatal is compatible with `log`.
func (log Logger) Fatal(arg0 interface{}, args ...interface{}) {
	log.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalln is compatible with `log`.
func (log Logger) Fatalln(arg0 interface{}, args ...interface{}) {
	log.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalf is compatible with `log`.
func (log Logger) Fatalf(format string, v ...interface{}) {
	log.intLog(ERROR, fmt.Sprintf(format, v...))
	os.Exit(0)
}

// Print is compatible with `log`.
func (log Logger) Print(arg0 interface{}, args ...interface{}) {
	log.intLog(INFO, arg0, args...)
}

// Println is compatible with `log`.
func (log Logger) Println(arg0 interface{}, args ...interface{}) {
	log.intLog(INFO, arg0, args...)
}

// Printf is compatible with `log`.
func (log Logger) Printf(format string, v ...interface{}) {
	log.intLog(INFO, fmt.Sprintf(format, v...))
}

// Panic is compatible with `log`.
func Panic(arg0 interface{}, args ...interface{}) {
	msg := FormatMessage(arg0, args...)
	std.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicln is compatible with `log`.
func Panicln(arg0 interface{}, args ...interface{}) {
	msg := FormatMessage(arg0, args...)
	std.intLog(CRITICAL, msg)
	panic(msg)
}

// Panicf is compatible with `log`.
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.intLog(CRITICAL, s)
	panic(s)
}

// Fatal is compatible with `log`.
func Fatal(arg0 interface{}, args ...interface{}) {
	std.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalln is compatible with `log`.
func Fatalln(arg0 interface{}, args ...interface{}) {
	std.intLog(ERROR, arg0, args...)
	os.Exit(0)
}

// Fatalf is compatible with `log`.
func Fatalf(format string, v ...interface{}) {
	std.intLog(ERROR, fmt.Sprintf(format, v...))
	os.Exit(0)
}

// Print is compatible with `log`.
func Print(arg0 interface{}, args ...interface{}) {
	std.intLog(INFO, arg0, args...)
}

// Println is compatible with `log`.
func Println(arg0 interface{}, args ...interface{}) {
	std.intLog(INFO, arg0, args...)
}

// Printf is compatible with `log`.
func Printf(format string, v ...interface{}) {
	std.intLog(INFO, fmt.Sprintf(format, v...))
}
