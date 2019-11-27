// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"net"
	"net/url"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

// Appender is an Appender that sends output to an UDP/TCP server
type Appender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout // format record for output
	sock   net.Conn
	prot   string
	host   string
}

// Close the socket if it opened.
func (a *Appender) Close() {
	if a.sock != nil {
		a.sock.Close()
	}
}

func init() {
	l4g.Register("socket", &Appender{})
}

// NewAppender creates a new appender with given
// socket protocol and endpoint.
func NewAppender(protocol, host string) *Appender {
	return &Appender{
		layout: l4g.NewPatternLayout(l4g.PatternJSON),
		sock:   nil,
		prot:   protocol,
		host:   host,
	}
}

// Open creates a socket Appender with DSN.
func (*Appender) Open(s string, args ...interface{}) (l4g.Appender, error) {
	protocol, address := "udp", "127.0.0.1:12124"
	if s != "" {
		if u, err := url.Parse(s); err == nil {
			if u.Scheme != "" {
				protocol = u.Scheme
			}
			if u.Host != "" {
				address = u.Host
			}
		}
	}
	return NewAppender(protocol, address).Set(args...), nil
}

// Write a log recorder to a socket.
// Connecting to the server on demand.
func (a *Appender) Write(e *l4g.Entry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.sock == nil {
		a.sock, err = net.Dial(a.prot, a.host)
		if err != nil {
			l4g.LogLogError(err)
			return
		}
	}

	_, err = a.sock.Write(a.layout.Format(e))
	if err != nil {
		l4g.LogLogError(err)
		a.sock.Close()
		a.sock = nil
	}
}

// Set options.
// Return Appender interface.
func (a *Appender) Set(args ...interface{}) l4g.Appender {
	ops, index, _ := l4g.ArgsToMap(args)
	for _, k := range index {
		a.SetOption(k, ops[k])
	}
	return a
}

// SetOption sets option with:
//	protocol - The named network. See net.Dial()
//	endpoint - The address and post number. See net.Dial()
//	pattern	 - Layout format pattern
//	utc 	 - Log recorder time zone
// Return errors
func (a *Appender) SetOption(k string, v interface{}) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	err = nil

	switch k {
	case "protocol":
		protocol := ""
		if protocol, err = l4g.ToString(v); err == nil && len(protocol) > 0 {
			a.Close()
			a.prot = protocol
		} else {
			err = l4g.ErrBadValue
		}
	case "endpoint":
		endpoint := ""
		if endpoint, err = l4g.ToString(v); err == nil && len(endpoint) > 0 {
			a.Close()
			a.host = endpoint
		} else {
			err = l4g.ErrBadValue
		}
	default:
		return a.layout.SetOption(k, v)
	}
	return
}
