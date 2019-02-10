package main

import (
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/socket"
	"os"
	"time"
)

var log = l4g.NewLogger(l4g.DEBUG)

func main() {
	// Enable internal log
	l4g.GetLogLog().Set("level", l4g.WARNING)

	fs := l4g.NewFilters().Add("network", l4g.FINEST, socketlog.NewSocketAppender("udp", "127.0.0.1:12124"))
	defer func() {
		if fs := log.GetFilters(); fs != nil {
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
}
