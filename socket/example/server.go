package main

import (
	"encoding/json"
	"flag"
	"fmt"
	l4g "github.com/ccpaging/nxlog4go"
	"net"
	"os"
)

var (
	port = flag.String("p", "12124", "Port number to listen on")
)

func e(err error) {
	if err != nil {
		fmt.Printf("Erroring out: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	// Bind to the port
	bind, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+*port)
	e(err)

	// Create listener
	listener, err := net.ListenUDP("udp", bind)
	e(err)

	var rec l4g.LogRecord
	fmt.Printf("Listening to port %s...\n", *port)

	for {
		// read into a new buffer
		buffer := make([]byte, 1024)
		size, a, err := listener.ReadFrom(buffer)
		e(err)

		// log to standard output
		fmt.Println(a, string(buffer[:size]))
		// fmt.Println(buffer[:size])
		err = json.Unmarshal(buffer[:size], &rec)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Decode:", rec)
		}
		fmt.Println("---")
	}
}
