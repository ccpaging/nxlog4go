// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/cast"
)

// Appender is an Appender that sends output to an UDP/TCP server
type Appender struct {
	mu       sync.Mutex         // ensures atomic writes; protects the following fields
	rec      chan *l4g.Recorder // entry channel
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

	l4g.Register("socket", &Appender{})
}

// NewAppender creates a socket appender with proto and hostport.
func NewAppender(proto, hostport string) *Appender {
	return &Appender{
		rec: make(chan *l4g.Recorder, 32),

		level:  l4g.INFO,
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

// Enabled encodes log Recorder and output it.
func (a *Appender) Enabled(r *l4g.Recorder) bool {
	if r == nil {
		return false
	}

	if r.Level < a.level {
		return false
	}

	a.runOnce.Do(func() {
		a.waitExit = &sync.WaitGroup{}
		a.waitExit.Add(1)
		go a.run(a.waitExit)
	})

	// Write after closed
	if a.waitExit == nil {
		a.output(r)
		return false
	}

	a.rec <- r
	return false
}

// Write is the filter's output method. This will block if the output
// buffer is full.
func (a *Appender) Write(b []byte) (int, error) {
	return 0, nil
}

func (a *Appender) run(waitExit *sync.WaitGroup) {
	for {
		select {
		case r, ok := <-a.rec:
			if !ok {
				waitExit.Done()
				return
			}
			a.output(r)
		}
	}
}

func (a *Appender) closeChannel() {
	// notify closing. See run()
	close(a.rec)
	// waiting for running channel closed
	a.waitExit.Wait()
	a.waitExit = nil
	// drain channel
	for r := range a.rec {
		a.output(r)
	}
}

// Close the socket if it opened.
func (a *Appender) Close() {
	if a.waitExit == nil {
		return
	}
	a.closeChannel()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sock != nil {
		a.sock.Close()
	}
}

// Output a log recorder to a socket. Connecting to the server on demand.
func (a *Appender) output(r *l4g.Recorder) {
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

	a.layout.Encode(buf, r)

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
//  level    - The output level
//
// Pattern layout options:
//	fromat	 - Layout format pattern
//  ...
//
// Return error
func (a *Appender) SetOption(k string, v interface{}) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var s string

	switch k {
	case "level":
		if _, ok := v.(int); ok {
			a.level = v.(int)
		} else if _, ok := v.(string); ok {
			a.level = l4g.Level(0).Int(v.(string))
		} else {
			err = fmt.Errorf("can not set option name %s, value %#v of type %T", k, v, v)
		}
	case "protocol": // DEPRECATED. See Open function's dsn argument
		if s, err = cast.ToString(v); err == nil && len(s) > 0 {
			if a.sock != nil {
				a.sock.Close()
			}
			a.proto = s
		}
	case "endpoint": // DEPRECATED. See Open function's dsn argument
		if s, err = cast.ToString(v); err == nil && len(s) > 0 {
			if a.sock != nil {
				a.sock.Close()
			}
			a.hostport = s
		}
	default:
		return a.layout.SetOption(k, v)
	}
	return
}
