// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"bytes"
	"net"
	"net/url"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

// Appender is an Appender that sends output to an UDP/TCP server
type Appender struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	layout l4g.Layout // format record for output

	proto    string
	hostport string
	sock     net.Conn
}

// Close the socket if it opened.
func (a *Appender) Close() {
	if a.sock != nil {
		a.sock.Close()
	}
}

/* Bytes Buffer */
var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	l4g.Register("socket", &Appender{})
}

// NewAppender creates a sock appender with proto and hostport.
func NewAppender(proto, hostport string) *Appender {
	return &Appender{
		layout: l4g.NewJSONLayout(),

		proto:    proto,
		hostport: hostport,
	}
}

// Open creates a socket Appender with DSN.
func (*Appender) Open(dsn string, args ...interface{}) (l4g.Appender, error) {
	proto, hostport := "udp", "127.0.0.1:12124"
	if dsn != "" {
		if u, err := url.Parse(dsn); err == nil {
			if u.Scheme != "" {
				proto = u.Scheme
			}
			if u.Host != "" {
				hostport = u.Host
			}
		}
	}
	return NewAppender(proto, hostport).Set(args...), nil
}

// Write a log recorder to a socket.
// Connecting to the server on demand.
func (a *Appender) Write(e *l4g.Entry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.sock == nil {
		a.sock, err = net.Dial(a.proto, a.hostport)
		if err != nil {
			l4g.LogLogError(err)
			return
		}
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	a.layout.Encode(buf, e)

	_, err = a.sock.Write(buf.Bytes())
	if err != nil {
		l4g.LogLogError(err)
		a.sock.Close()
		a.sock = nil
	}
}

// Set options.
// Return Appender interface.
func (a *Appender) Set(args ...interface{}) l4g.Appender {
	ops, idx, _ := l4g.ArgsToMap(args)
	for _, k := range idx {
		a.SetOption(k, ops[k])
	}
	return a
}

// SetOption sets option with:
//	pattern	 - Layout format pattern
//  ...
//
// Return error
func (a *Appender) SetOption(k string, v interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch k {
	case "protocol": // DEPRECATED. See Open function's dsn argument
		if proto, err := l4g.ToString(v); err == nil && len(proto) > 0 {
			if a.sock != nil {
				a.sock.Close()
			}
			a.proto = proto
		} else {
			return l4g.ErrBadValue
		}
	case "endpoint": // DEPRECATED. See Open function's dsn argument
		if hostport, err := l4g.ToString(v); err == nil && len(hostport) > 0 {
			if a.sock != nil {
				a.sock.Close()
			}
			a.hostport = hostport
		} else {
			return l4g.ErrBadValue
		}
	default:
		return a.layout.SetOption(k, v)
	}
	return nil
}
