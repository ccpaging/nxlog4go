// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"fmt"
	"net"
	"os"
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

func (sa *SocketAppender) Close() {
	if sa.sock != nil {
		sa.sock.Close()
	}
}

func NewAppender(prot, host string) *SocketAppender {
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
			fmt.Fprintf(os.Stderr, "SocketAppender(%s): %v\n", sa.host, err)
			return
		}
	}

	_, err = sa.sock.Write(sa.layout.Format(rec))
	if err != nil {
		fmt.Fprintf(os.Stderr, "SocketAppender(%s): %v\n", sa.host, err)
		sa.sock.Close()
		sa.sock = nil
	}
}

// Set option. chainable
func (sa *SocketAppender) Set(name string, v interface{}) *SocketAppender {
	sa.SetOption(name, v)
	return sa
}

// Set option. checkable
func (sa *SocketAppender) SetOption(name string, v interface{}) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	switch name {
	case "protocol":
		if protocol, ok := v.(string); ok {
			sa.Close()
			sa.prot = protocol
		} else {
			return l4g.ErrBadValue
		}
	case "endpoint":
		if endpoint, ok := v.(string); ok {
			if len(endpoint) > 0 {
				sa.Close()
				sa.host = endpoint
			} else {
				return l4g.ErrBadValue
			}
		} else {
			return l4g.ErrBadValue
		}
	case "pattern", "utc":
		return sa.layout.SetOption(name, v)
	default:
		return l4g.ErrBadOption
	}
	return nil
}
