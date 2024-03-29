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
	"github.com/ccpaging/nxlog4go/patt"
)

// Appender is an Appender that sends output to an UDP/TCP server
type Appender struct {
	mu       sync.Mutex            // ensures atomic writes; protects the following fields
	rec      chan *driver.Recorder // entry channel
	runOnce  sync.Once
	waitExit *sync.WaitGroup

	level  int
	layout driver.Layout // format entry for output

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

	driver.Register("socket", &Appender{})
}

// NewAppender creates a socket appender with proto and hostport.
func NewAppender(proto, hostport string) *Appender {
	return &Appender{
		rec: make(chan *driver.Recorder, 32),

		layout: patt.NewJSONLayout(),

		proto:    proto,
		hostport: hostport,
	}
}

// Open creates an Appender with DSN.
func (*Appender) Open(dsn string, args ...interface{}) (driver.Appender, error) {
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
	return NewAppender(proto, hostport).SetOptions(args...), nil
}

// Layout returns the output layout for the appender.
func (sa *Appender) Layout() driver.Layout {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	return sa.layout
}

// SetLayout sets the output layout for the appender.
func (sa *Appender) SetLayout(layout driver.Layout) *Appender {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.layout = layout
	return sa
}

// SetOptions sets name-value pair options.
//
// Return the appender.
func (sa *Appender) SetOptions(args ...interface{}) *Appender {
	ops, idx, _ := driver.ArgsToMap(args...)
	for _, k := range idx {
		sa.Set(k, ops[k])
	}
	return sa
}

// Enabled encodes log Recorder and output it.
func (sa *Appender) Enabled(r *driver.Recorder) bool {
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
func (sa *Appender) Write(b []byte) (int, error) {
	return 0, nil
}

func (sa *Appender) run(waitExit *sync.WaitGroup) {
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

func (sa *Appender) closeChannel() {
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
func (sa *Appender) Close() {
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
func (sa *Appender) output(r *driver.Recorder) {
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

// Set sets name-value option with:
//  level    - The output level
//
// Pattern layout options:
//	pattern	 - Layout format pattern
//  ...
//
// Return error
func (sa *Appender) Set(k string, v interface{}) (err error) {
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
