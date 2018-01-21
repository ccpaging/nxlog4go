// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package socketlog

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	l4g "github.com/ccpaging/nxlog4go"
)

// This log writer sends output to a socket
type SocketLogWriter struct {
	mu   sync.Mutex // ensures atomic writes; protects the following fields
	sock net.Conn
	prot string
	host string
}

func (slw *SocketLogWriter) Close() {
	if slw.sock != nil {
		slw.sock.Close()
	}
}

func NewLogWriter(prot, host string) *SocketLogWriter {
	return &SocketLogWriter {
		sock:	nil,
		prot:	prot,
		host:	host,
	}
}

func (slw *SocketLogWriter) LogWrite(rec *l4g.LogRecord) {
	slw.mu.Lock()
	defer slw.mu.Unlock()

	js, err := json.Marshal(rec)
	if err != nil {
		fmt.Fprint(os.Stderr, "SocketLogWriter(%s): %s", slw.host, err)
		return
	}
	js = append(js, '\n')

	if slw.sock == nil {
		slw.sock, err = net.Dial(slw.prot, slw.host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SocketLogWriter(%s): %v\n", slw.host, err)
			return
		}
	}

	_, err = slw.sock.Write(js)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SocketLogWriter(%s): %v\n", slw.host, err)
		slw.sock.Close()
		slw.sock = nil
	}
}

// Set option. chainable
func (slw *SocketLogWriter) Set(name string, v interface{}) *SocketLogWriter {
	slw.SetOption(name, v)
	return slw
}

// Set option. checkable
func (slw *SocketLogWriter) SetOption(name string, v interface{}) error {
	slw.mu.Lock()
	defer slw.mu.Unlock()

	switch name {
	case "protocol":
		if protocol, ok := v.(string); ok {
			slw.Close()
			slw.prot = protocol
		} else {
			return l4g.ErrBadValue
		}
	case "endpoint":
		if endpoint, ok := v.(string); ok {
			if len(endpoint) > 0 {
				slw.Close()
				slw.host = endpoint
			} else {
				return l4g.ErrBadValue
			}
		} else {
			return l4g.ErrBadValue
		}
	default:
		return l4g.ErrBadOption
	}
	return nil
}
