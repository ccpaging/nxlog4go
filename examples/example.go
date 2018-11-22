package main

import (
	"time"
	"os"
	"io/ioutil"

	l4g "github.com/ccpaging/nxlog4go"
)


var glog = l4g.New(l4g.DEBUG).Set("prefix", "example").Set("pattern", "[%T %D %Z] [%L] (%P:%s) %M\n")

func main() {
	glog.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	log1 := l4g.NewLogger(os.Stderr, l4g.DEBUG, "prefix1", "[%N %D %z] [%L] (%P:%s) %M\n")
	log1.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	// set io.Writer as nil, no output
	log2 := l4g.New(l4g.DEBUG).SetOutput(ioutil.Discard)
	log2.Info("Write to Discard. The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	// level filter, no output
	log3 := l4g.New(l4g.INFO)
	log3.Debug("Filter out. The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	// change time zone to 0
	glog.Layout().Set("utc", true)
	glog.Info("Using UTC time stamp. Now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Layout().Set("utc", false)
	glog.Info("Using local time stamp. Now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
}
