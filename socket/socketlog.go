// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"net"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

// This log appender sends output to a socket
type SocketAppender struct {
	mu   sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout 	 // format record for output
	sock net.Conn
	prot string
	host string
}

func (sa *SocketAppender) Init() {
}

func (sa *SocketAppender) Close() {
	if sa.sock != nil {
		sa.sock.Close()
	}
}

func NewAppender(prot, host string) l4g.Appender {
	return &SocketAppender {
		layout: l4g.NewPatternLayout(l4g.PATTERN_JSON),	
		sock:	nil,
		prot:	prot,
		host:	host,
	}
}

// This is the SocketAppender's output method
func (sa *SocketAppender) Write(rec *l4g.LogRecord) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	var err error
	if sa.sock == nil {
		sa.sock, err = net.Dial(sa.prot, sa.host)
		if err != nil {
			l4g.LogLogError("SocketAppender", err)
			return
		}
	}

	_, err = sa.sock.Write(sa.layout.Format(rec))
	if err != nil {
		l4g.LogLogError("SocketAppender", err)
		sa.sock.Close()
		sa.sock = nil
	}
}

// Set option. chainable
func (sa *SocketAppender) Set(name string, v interface{}) l4g.Appender {
	sa.SetOption(name, v)
	return sa
}

/* 
Set option. checkable. Better be set before SetFilters()
Option names include:
	protocol - The named network. See net.Dial()
	endpoint - The address and post number. See net.Dial()
	pattern	 - Layout format pattern
	utc 	 - Log recorder time zone
*/
func (sa *SocketAppender) SetOption(name string, v interface{}) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	switch name {
	case "protocol":
		if protocol, ok := v.(string); !ok {
			return l4g.ErrBadValue
		} else if protocol == "" {
			return l4g.ErrBadValue
		} else {
			sa.Close()
			sa.prot = protocol
		}
	case "endpoint":
		if endpoint, ok := v.(string); !ok {
			return l4g.ErrBadValue
		} else if endpoint == "" {
			return l4g.ErrBadValue
		} else {
			sa.Close()
			sa.host = endpoint
		}
	default:
		return sa.layout.SetOption(name, v)
	}
	return nil
}
