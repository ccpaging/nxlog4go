// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	l4g "github.com/ccpaging/nxlog4go"
)

// JSONAppender represents the log appender that sends JSON format records to a file
type JSONAppender struct {
	*Appender
}

func init() {
	l4g.Register("json", &JSONAppender{})
}

// Open creates a new file appender which json format.
func (*JSONAppender) Open(filename string, args ...interface{}) (l4g.Appender, error) {
	a, err := NewAppender(filename, args...)
	if err != nil {
		return nil, err
	}
	a.layout = l4g.NewJSONLayout(args...)
	return &JSONAppender{a}, nil
}
