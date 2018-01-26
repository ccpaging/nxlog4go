package main

import (
	"time"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/socket"
)

var	log = l4g.New(l4g.DEBUG)

func main() {
	// This makes sure the output stream buffer is written
	sa := socketlog.NewAppender("udp", "127.0.0.1:12124")
	// defer slw.Close()

	fs := l4g.NewFilters().Add("network", l4g.FINEST, sa)
	defer fs.Close()
	log.SetFilters(fs)

	// Run `nc -u -l -p 12124` or similar before you run this to see the following message
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	}
	log.SetFilters(nil)
	sa.Close()
}
