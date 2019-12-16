package main

import (
	"encoding/json"
	"flag"
	"fmt"
	l4g "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/console"
	_ "github.com/ccpaging/nxlog4go/file"
	_ "github.com/ccpaging/nxlog4go/socket"
	"os"
)

var (
	debug = flag.Bool("debug", false, "")
	fname = flag.String("conf", "example.json", "config file")
)

var log = l4g.GetLogger().SetOptions("caller", false, "format", "[%T] [%L] (%S) %M\n")

func main() {
	flag.Parse()

	// Enable internal logger
	l4g.GetLogLog().Set("level", l4g.TRACE)

	// Open config file
	fd, err := os.Open(*fname)
	if err != nil {
		panic(fmt.Sprintf("Can't load json config file: %s %v", fname, err))
	}

	lc := new(l4g.LoggerConfig)
	// decode json
	err = json.NewDecoder(fd).Decode(&lc)
	fd.Close()
	if err != nil {
		panic(fmt.Sprintf("Can't parse json config file: %s %v", fname, err))
	}
	fmt.Printf("Total configuration: %d\n", len(lc.Filters))

	// Configure logger
	errs := log.LoadConfiguration(lc)
	for _, err := range errs {
		fmt.Println(err)
	}
	fmt.Printf("Total appenders installed: %d\n", len(log.Filters()))

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Trace("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	// Close all appenders in logger
	log.Close()

	os.Remove("_test.log")
	os.Remove("_trace.xml")
}
