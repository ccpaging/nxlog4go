package main

import (
	"time"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/socket"
)

var	log = l4g.New(l4g.DEBUG)

func main() {
	// This makes sure the output stream buffer is written
	slw := socketlog.NewLogWriter().Set("protocol", "udp").Set("endpoint", "127.0.0.1:12124")
	// defer slw.Close()

	log.AddFilter("network", l4g.FINEST, slw)
	defer log.CloseFilters()

	// Run `nc -u -l -p 12124` or similar before you run this to see the following message
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	}
}
