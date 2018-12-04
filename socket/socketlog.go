// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"net"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

// SocketAppender is an Appender that sends output to an UDP/TCP server
type SocketAppender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout // format record for output
	sock   net.Conn
	prot   string
	host   string
}

// Init is nothing to do here.
func (sa *SocketAppender) Init() {
}

// Close the socket if it opened.
func (sa *SocketAppender) Close() {
	if sa.sock != nil {
		sa.sock.Close()
	}
}

func init() {
	l4g.AddAppenderNewFunc("socket", New)
}

// New creates a socket Appender with default udp protocol and endpoint.
func New() l4g.Appender {
	return NewSocketAppender("udp", "127.0.0.1:12124")
}

// NewSocketAppender creates a new appender with given
// socket protocol and endpoint.
func NewSocketAppender(prot, host string) l4g.Appender {
	return &SocketAppender{
		layout: l4g.NewPatternLayout(l4g.PatternJSON),
		sock:   nil,
		prot:   prot,
		host:   host,
	}
}

// Write a log recorder to a socket.
// Connecting to the server on demand.
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

// Set option.
// Return Appender interface.
func (sa *SocketAppender) Set(k string, v interface{}) l4g.Appender {
	sa.SetOption(k, v)
	return sa
}

// SetOption sets option with:
//	protocol - The named network. See net.Dial()
//	endpoint - The address and post number. See net.Dial()
//	pattern	 - Layout format pattern
//	utc 	 - Log recorder time zone
// Return errors
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
