// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package file

import (
	"github.com/ccpaging/nxlog4go/driver"
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
	*Appender
}

func init() {
	driver.Register("xml", &XMLAppender{})
}

// NewXMLAppender creates a new file appender with recorder XML format.
func NewXMLAppender(filename string, args ...interface{}) (*XMLAppender, error) {
	a, err := NewAppender(filename, args...)
	if err != nil {
		return nil, err
	}
	a.SetOptions("head", XMLHead, "format", XMLRecord, "foot", XMLFoot)
	return &XMLAppender{a}, nil
}

// Open creates a new file appender which XML format.
func (*XMLAppender) Open(filename string, args ...interface{}) (driver.Appender, error) {
	return NewXMLAppender(filename, args...)
}
