package main

import (
	"encoding/xml"
	"os"
	"fmt"
	"bufio"
	"io"
	"io/ioutil"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/file"
	"github.com/ccpaging/nxlog4go/socket"
)

var filename string = "config.xml"

var log = l4g.GetLogger()

// Print what was logged to the file (yes, I know I'm skipping error checking)
func PrintFile(fn string) {
	fd, _ := os.Open(fn)
	in := bufio.NewReader(fd)
	fmt.Print("Messages logged to file were: (line numbers not included)\n")
	for lineno := 1; ; lineno++ {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Printf("%3d:\t%s", lineno, line)
	}
	fd.Close()
}

func main() {
	// disable default console writer
	log.SetOutput(nil)

	// Open config file
	fd, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Can't load xml config file: %s %v", filename, err))
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read %q: %s\n", filename, err)
		os.Exit(1)
	}

	fd.Close()

	xc := new(l4g.LoggerConfig)
	if err := xml.Unmarshal(buf, xc); err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse XML configuration. %s\n", err)
		os.Exit(1)
	}

	fs := l4g.NewFilters()
	var	lw l4g.LogWriter
	for _, fc := range xc.FilterConfigs {
		bad, enabled, lvl := l4g.CheckFilterConfig(fc)
	
		if bad {
			os.Exit(1)
		}
	
		switch fc.Type {
		case "color":
			lw = colorlog.NewLogWriter()
		case "file":
			lw = filelog.NewLogWriter("_test.log", 0)
		case "socket":
			lw = socketlog.NewLogWriter("udp", "127.0.0.1:12124")
		case "xml":
			lw = filelog.NewLogWriter("_test.log", 0)
			lw.SetOption("head","<log created=\"%D %T\">")
			lw.SetOption("format", 
`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>`)
			lw.SetOption("foot", "</log>")
		default:
			panic(fmt.Sprintf("Unknown filter type \"%s\"", fc.Type))
		}
	
		if lw == nil {
			panic(fmt.Sprintf("Unknown filter type \"%s\"", fc.Type))
		}

		good := l4g.SetLogWriter(lw, fc.Properties)
		if !good {
			fmt.Println(fc.Tag, "NOT good")
			continue
		}
	
		// If it's disabled, we're just checking syntax
		if !enabled {
			fmt.Println(fc.Tag, "disabled")
			continue
		}
		
		fmt.Println("Add filter", fc.Tag, lvl)
		fs.Add(fc.Tag, lvl, lw)
	}

	log.SetFilters(fs)
	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Trace("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	log.SetFilters(nil)
	fs.Close()

	PrintFile("_test.log")
	os.Remove("_test.log")
	PrintFile("_trace.xml")
	os.Remove("_trace.xml")
}

