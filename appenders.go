// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
)

/****** Appenders map ******/

// AppenderNewFunc represents a function that create a new appender interface
type AppenderNewFunc func() Appender

var appenders = make(map[string]AppenderNewFunc)

// AddAppenderNewFunc is called by 3rd appender's init() function 
// to register New() function that creates and returns Appender interface.
func AddAppenderNewFunc(typ string, newFunc AppenderNewFunc) {
	if typ == "" {
		return
	}
	if newFunc == nil {
		delete(appenders, typ)
		return
	}
	appenders[typ] = newFunc
}

func getAppenderNewFunc(typ string) AppenderNewFunc {
	if newFunc, ok := appenders[typ]; ok {
		return newFunc
	}

	return nil
}
