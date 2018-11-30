// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"net"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

// This log appender sends output to a socket
type SocketAppender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout // format record for output
	sock   net.Conn
	prot   string
	host   string
}

func (sa *SocketAppender) Init() {
}

func (sa *SocketAppender) Close() {
	if sa.sock != nil {
		sa.sock.Close()
	}
}

func init() {
	l4g.AddAppenderNewFunc("socket", New)
}

// This creates a the socket appender with default udp
// protocol and endpoint.
func New() l4g.Appender {
	return NewSocketAppender("udp", "127.0.0.1:12124")
}

// NewSocketAppender creates a new appender with given
// socket protocol and endpoint.
func NewSocketAppender(prot, host string) l4g.Appender {
	return &SocketAppender{
		layout: l4g.NewPatternLayout(l4g.PatternJson),
		sock:   nil,
		prot:   prot,
		host:   host,
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
			l4g.LogLogError(err)
			return
		}
	}

	_, err = sa.sock.Write(sa.layout.Format(rec))
	if err != nil {
		l4g.LogLogError(err)
		sa.sock.Close()
		sa.sock = nil
	}
}

// Set option. chainable
func (sa *SocketAppender) Set(k string, v interface{}) l4g.Appender {
	sa.SetOption(k, v)
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
func (sa *SocketAppender) SetOption(k string, v interface{}) (err error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	err = nil

	switch k {
	case "protocol":
		protocol := ""
		if protocol, err = l4g.ToString(v); err == nil && len(protocol) > 0 {
			sa.Close()
			sa.prot = protocol
		} else {
			err = l4g.ErrBadValue
		}
	case "endpoint":
		endpoint := ""
		if endpoint, err = l4g.ToString(v); err == nil && len(endpoint) > 0 {
			sa.Close()
			sa.host = endpoint
		} else {
			err = l4g.ErrBadValue
		}
	default:
		return sa.layout.SetOption(k, v)
	}
	return
}
