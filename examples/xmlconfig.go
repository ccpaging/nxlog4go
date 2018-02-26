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

	// pre-install appender may used in configuration
	fs.Add("color", l4g.OFFLevel, colorlog.NewAppender())
	fs.Add("file", l4g.OFFLevel, filelog.NewAppender("_test.log", 0))
	fs.Add("socket", l4g.OFFLevel, socketlog.NewAppender("udp", "127.0.0.1:12124"))
	xa := filelog.NewAppender("_test.log", 0)
	xa.SetOption("head","<log created=\"%D %T\">%R")
	xa.SetOption("pattern", 
`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>%R`)
	xa.SetOption("foot", "</log>%R")
	fs.Add("xml", l4g.OFFLevel, xa)
	
	fmt.Println(len(*fs), "appenders pre-installed")
	fs.LoadConfiguration(xc.Filters)
	if len(*fs) > 0 {
		log.SetFilters(fs)
		// disable default console writer
		log.SetOutput(nil)
	}
	fmt.Println(len(*fs), "appenders configured ok")

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

