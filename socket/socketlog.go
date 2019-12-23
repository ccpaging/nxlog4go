// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"bytes"
	"net"
	"net/url"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/cast"
	"github.com/ccpaging/nxlog4go/driver"
)

// SocketAppender is an Appender that sends output to an UDP/TCP server
type SocketAppender struct {
	mu       sync.Mutex            // ensures atomic writes; protects the following fields
	rec      chan *driver.Recorder // entry channel
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	level  int
	layout l4g.Layout // format entry for output

	proto    string
	hostport string
	sock     net.Conn
}

/* Bytes Buffer */
var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	driver.Register("socket", &SocketAppender{})
}

// NewSocketAppender creates a socket appender with proto and hostport.
func NewSocketAppender(proto, hostport string) *SocketAppender {
	return &SocketAppender{
		rec: make(chan *driver.Recorder, 32),

		layout: l4g.NewJSONLayout(),

		proto:    proto,
		hostport: hostport,
	}
}

// Open creates an Appender with DSN.
func (*SocketAppender) Open(dsn string, args ...interface{}) (driver.Appender, error) {
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
	return NewSocketAppender(proto, hostport).SetOptions(args...), nil
}

// Enabled encodes log Recorder and output it.
func (sa *SocketAppender) Enabled(r *driver.Recorder) bool {
	// r.Level < fa.level
	if !(r.Level >= sa.level) {
		return false
	}

	sa.runOnce.Do(func() {
		sa.waitExit = &sync.WaitGroup{}
		sa.waitExit.Add(1)
		go sa.run(sa.waitExit)
	})

	// Write after closed
	if sa.waitExit == nil {
		sa.output(r)
		return false
	}

	sa.rec <- r
	return false
}

// Write is the filter's output method. This will block if the output
// buffer is full.
func (sa *SocketAppender) Write(b []byte) (int, error) {
	return 0, nil
}

func (sa *SocketAppender) run(waitExit *sync.WaitGroup) {
	for {
		select {
		case r, ok := <-sa.rec:
			if !ok {
				waitExit.Done()
				return
			}
			sa.output(r)
		}
	}
}

func (sa *SocketAppender) closeChannel() {
	// notify closing. See run()
	close(sa.rec)
	// waiting for running channel closed
	sa.waitExit.Wait()
	sa.waitExit = nil
	// drain channel
	for r := range sa.rec {
		sa.output(r)
	}
}

// Close the socket if it opened.
func (sa *SocketAppender) Close() {
	if sa.waitExit == nil {
		return
	}
	sa.closeChannel()

	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.sock != nil {
		sa.sock.Close()
	}
}

// Output a log recorder to a socket. Connecting to the server on demand.
func (sa *SocketAppender) output(r *driver.Recorder) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	var err error
	if sa.sock == nil {
		sa.sock, err = net.Dial(sa.proto, sa.hostport)
		if err != nil {
			l4g.LogLogError(err)
			return
		}
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	sa.layout.Encode(buf, r)

	_, err = sa.sock.Write(buf.Bytes())
	if err != nil {
		l4g.LogLogError(err)
		sa.sock.Close()
		sa.sock = nil
	}
}

// SetOptions sets name-value pair options.
//
// Return Appender interface.
func (sa *SocketAppender) SetOptions(args ...interface{}) *SocketAppender {
	ops, idx, _ := driver.ArgsToMap(args)
	for _, k := range idx {
		sa.Set(k, ops[k])
	}
	return sa
}

// Set sets name-value option with:
//  level    - The output level
//
// Pattern layout options:
//	fromat	 - Layout format pattern
//  ...
//
// Return error
func (sa *SocketAppender) Set(k string, v interface{}) (err error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	var s string

	switch k {
	case "level":
		var n int
		if n, err = l4g.Level(l4g.INFO).IntE(v); err == nil {
			sa.level = n
		}
	case "protocol": // DEPRECATED. See Open function's dsn argument
		if s, err = cast.ToString(v); err == nil && len(s) > 0 {
			if sa.sock != nil {
				sa.sock.Close()
			}
			sa.proto = s
		}
	case "endpoint": // DEPRECATED. See Open function's dsn argument
		if s, err = cast.ToString(v); err == nil && len(s) > 0 {
			if sa.sock != nil {
				sa.sock.Close()
			}
			sa.hostport = s
		}
	default:
		return sa.layout.Set(k, v)
	}
	return
}
