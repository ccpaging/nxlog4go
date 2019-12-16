package main

import (
	"io/ioutil"
	"os"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
)

var glog = l4g.NewLogger(l4g.DEBUG).SetOptions("prefix", "example", "format", "[%T %D %Z] [%L] (%P:%S) %M")

var isTTY bool
var out = os.Stderr

func isatty(file *os.File) bool {
	// This is sort of cheating: if stdout is a character device, we assume
	// that means it's a TTY. Unfortunately, there are many non-TTY
	// character devices, but fortunately stdout is rarely set to any of
	// them.
	//
	// We could solve this properly by pulling in a dependency on
	// code.google.com/p/go.crypto/ssh/terminal, for instance, but as a
	// heuristic for whether to print in color or in black-and-white, I'd
	// really rather not.
	fi, err := file.Stat()
	if err == nil {
		m := os.ModeDevice | os.ModeCharDevice
		return (fi.Mode()&m == m)
	}
	return false
}

func init() {
	isTTY = isatty(out)
	isTTY = isTTY || (os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb")
	isTTY = isTTY || (os.Getenv("ConEmuANSI") == "ON")
}

func main() {
	glog.SetOutput(out).SetOptions("color", isTTY)
	glog.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Trace("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Error("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Critical("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	log1 := l4g.NewLogger(l4g.DEBUG).SetOptions("prefix", "example", "format", "%P[%T %D %Z] [%L] (%S:%N) %M")
	log1.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	// set io.Writer as nil, no output
	log2 := l4g.NewLogger(l4g.DEBUG).SetOutput(ioutil.Discard)
	log2.Info("Write to Discard. The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	// level filter, no output
	log3 := l4g.NewLogger(l4g.INFO)
	log3.Debug("Filter out. The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	// change time zone to 0
	glog.Layout().Set("utc", true)
	glog.Info("Using UTC time stamp. Now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	glog.Layout().Set("utc", false)
	glog.Info("Using local time stamp. Now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
}
