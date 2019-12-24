// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

// JSONAppender represents the log appender that sends JSON format records to a file
type JSONAppender struct {
	*FileAppender
}

func init() {
	driver.Register("json", &JSONAppender{})
}

// Open creates a new file appender which json format.
func (*JSONAppender) Open(filename string, args ...interface{}) (driver.Appender, error) {
	a, err := NewFileAppender(filename, args...)
	if err != nil {
		return nil, err
	}
	a.layout = patt.NewJSONLayout(args...)
	return &JSONAppender{a}, nil
}
