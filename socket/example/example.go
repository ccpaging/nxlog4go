package main

import (
	"os"
	"time"
	"net"
	"fmt"
	"encoding/json"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/socket"
)

var addr = "127.0.0.1:12124"

func e(err error) {
	if err != nil {
		fmt.Printf("Erroring out: %s\n", err)
		os.Exit(1)
	}
}

func server(ready chan struct{}) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	e(err)

	conn, err := net.ListenUDP("udp", laddr)
	e(err)
	defer conn.Close()
	
	var rec l4g.LogRecord
	fmt.Printf("Listening on %v...\n", laddr)

	close(ready)
	for {
		// read into a new buffer
		buffer := make([]byte, 1024)
		size, a, err := conn.ReadFrom(buffer)
		e(err)

		// log to standard output
		fmt.Println(a, string(buffer[:size]))
		// fmt.Println(buffer[:size])
		err = json.Unmarshal(buffer[:size], &rec)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Unmarshal:", rec)
		}
		fmt.Println("---")
	}
}

func client() {
	// Enable internal log
	l4g.GetLogLog().Set("level", l4g.WARNING)

	log := l4g.NewLogger(l4g.DEBUG).SetPrefix("client").Set("pattern", "%P " + l4g.PatternDefault)

	fs := l4g.NewFilters().Add("network", l4g.FINEST, socketlog.NewSocketAppender("udp", addr))
	defer func() {
		if fs := log.Filters(); fs != nil {
			log.SetFilters(nil).SetOutput(os.Stderr)
			fs.Close()
		}
	}()

	log.SetFilters(fs)

	// Run `nc -u -l -p 12124` or similar before you run this to see the following message
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	}
	log.Shutdown()
}

func main() {
	ready := make(chan struct{})
	go server(ready)
	<-ready
	
	client()
}
