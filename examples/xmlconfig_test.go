package main

import (
	"testing"

	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	l4g "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/console"
	_ "github.com/ccpaging/nxlog4go/file"
	_ "github.com/ccpaging/nxlog4go/socket"
)

func TestXMLConfig(t *testing.T) {
	var fname string = "example.xml"
	var log = l4g.GetLogger()

	// Enable internal logger
	l4g.GetLogLog().Set("level", l4g.TRACE)

	// Open config file
	fd, err := os.Open(fname)
	if err != nil {
		panic(fmt.Sprintf("Can't load xml config file: %s %v", fname, err))
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read %q: %s\n", fname, err)
		os.Exit(1)
	}

	fd.Close()

	lc := new(l4g.LoggerConfig)
	if err := xml.Unmarshal(buf, lc); err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse XML configuration. %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total configuration: %d\n", len(lc.Filters))
	// fmt.Println(lc)

	errs := log.LoadConfiguration(lc)
	for _, err := range errs {
		fmt.Println(err)
	}

	filters := log.Filters()

	fmt.Printf("Total appenders installed: %d\n", len(filters))

	// disable default console writer
	log.SetOutput(nil)

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Trace("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	// Do not forget close all filters
	for _, f := range filters {
		// Unload filters
		log.Detach(f)
		if f != nil {
			f.Close()
		}
	}

	os.Remove("_test.log")
	os.Remove("_trace.xml")
}
