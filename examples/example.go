package main

import (
	"time"
	"os"

	l4g "github.com/ccpaging/nxlog4go"
)


var glog = l4g.New(l4g.DEBUG).SetPrefix("example").SetFormat("[%T %D %Z] [%L] (%P:%s) %M")

func main() {
	glog.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	log1 := l4g.NewLogger(os.Stderr, l4g.DEBUG, "prefix1", "[%N %D %Z] [%L] (%P:%s) %M")
	log1.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	// set io.Writer as nil, no output
	log2 := l4g.New(l4g.DEBUG).SetOutput(nil)
	log2.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	// level filter, no output
	log3 := l4g.New(l4g.INFO)
	log3.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	time.Sleep(1 * time.Second)

	// change time zone to 0
	l4g.FORMAT_UTC = true
	glog.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
}
