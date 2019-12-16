// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	l4g "github.com/ccpaging/nxlog4go"
)

var (
	// XMLHead is layout format of log file header
	XMLHead = "<log created=\"%D %T\">"
	// XMLRecord is layout format of log record
	XMLRecord = "\t<record level=\"%L\">\n" +
		"\t\t<timestamp>%D %T</timestamp>\n" +
		"\t\t<source>%S</source>\n" +
		"\t\t<message>%M</message>\n" +
		"\t</record>"
	// XMLFoot is layout format of log file trailer
	XMLFoot = "</log>"
)

// XMLAppender represents the log appender that sends XML format records to a file
type XMLAppender struct {
	*FileAppender
}

func init() {
	l4g.Register("xml", &XMLAppender{})
}

// Open creates a new file appender which XML format.
func (*XMLAppender) Open(filename string, args ...interface{}) (l4g.Appender, error) {
	a, err := NewFileAppender(filename, args...)
	if err != nil {
		return nil, err
	}
	a.SetOptions("head", XMLHead, "format", XMLRecord, "foot", XMLFoot)
	return &XMLAppender{a}, nil
}
