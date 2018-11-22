// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
)

/****** Appenders map ******/

type AppenderNewFunc func() Appender

type Appenders map[string]AppenderNewFunc

var appenders = make(Appenders)

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

func GetAppenderNewFunc(typ string) AppenderNewFunc {
	if newFunc, ok := appenders[typ]; ok {
		return newFunc
	}

	return nil
}
