// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package filelog

import (
	l4g "github.com/ccpaging/nxlog4go"
)

var (
	// XMLHead is layout pattern of log file header
	XMLHead = "<log created=\"%D %T\">%R"
	// XMLPattern is layout pattern of log record
	XMLPattern = `	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>%R`
	// XMLHead is layout pattern of log file trailer
	XMLFoot = "</log>%R"
)

type XMLAppender struct {
	*Appender
}

func init() {
	l4g.Register("xml", &XMLAppender{})
}

// Open creates a new file appender which XML format.
func (*XMLAppender) Open(filename string, args ...interface{}) (l4g.Appender, error) {
	a, err := NewAppender(filename, args...)
	if err != nil {
		return nil, err
	}

	a.SetOption("head", XMLHead)
	a.SetOption("pattern", XMLPattern)
	a.SetOption("foot", XMLFoot)
	return &XMLAppender{a}, err
}
