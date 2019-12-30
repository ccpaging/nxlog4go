package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
	_ "github.com/ccpaging/nxlog4go/socket"
)

var addr = "127.0.0.1:12124"

func checkError(err error) {
	if err != nil {
		fmt.Printf("Erroring out: %s\n", err)
		os.Exit(1)
	}
}

func server(ready chan struct{}) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	checkError(err)

	conn, err := net.ListenUDP("udp", laddr)
	checkError(err)
	defer conn.Close()

	var r driver.Recorder
	fmt.Printf("Listening on %v...\n", laddr)

	close(ready)
	for {
		// read into a new buffer
		buffer := make([]byte, 1024)
		size, a, err := conn.ReadFrom(buffer)
		if size > 0 {
			// log to standard output
			fmt.Println(a, string(buffer[:size]))
			// fmt.Println(buffer[:size])
			err = json.Unmarshal(buffer[:size], &r)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Unmarshal:", r)
			}
			fmt.Println("---")
		}
		checkError(err)
	}
}

func client() {
	// Enable internal log
	l4g.GetLogLog().Set("level", l4g.WARN)

	log := l4g.NewLogger(l4g.DEBUG).SetPrefix("client").SetOptions("format", "%P "+patt.FormatDefault)

	sa, err := driver.Open("socket", "udp://"+addr, "level", l4g.FINEST)
	checkError(err)

	log.AddFilter("socket", 0, sa)
	defer func() {
		log.Close()
	}()

	// Run `nc -u -l -p 12124` or similar before you run this to see the following message
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	time.Sleep(1 * time.Second)

	for i := 0; i < 5; i++ {
		log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		time.Sleep(1 * time.Second)
	}
}

func main() {
	ready := make(chan struct{})
	go server(ready)
	<-ready

	client()
}
