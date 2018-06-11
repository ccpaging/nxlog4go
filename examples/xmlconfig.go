package main

import (
	"encoding/xml"
	"os"
	"fmt"
	"bufio"
	"io"
	"io/ioutil"
	"strings"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/file"
	"github.com/ccpaging/nxlog4go/socket"
)

var filename string = "config.xml"

var log = l4g.GetLogger()
var filters = l4g.NewFilters()

// Appender's properties
type Prop struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type FilterConfig struct {
	Tag	  string `xml:"tag"`
	Level string `xml:"level"`
	Props []Prop `xml:"property"`
}

type LoggerConfig struct {
	Filters  []FilterConfig `xml:"filter"`
}

func loadFilter(fc *FilterConfig) {
	if len(fc.Tag) == 0 {
		log.Error("Required child tag")
		return
	}
	tag := strings.ToLower(fc.Tag)
	if len(fc.Level) == 0 {
		log.Error("Required child level")
		return
	}
	lvl := l4g.GetLevel(fc.Level)
	if lvl >= l4g.SILENT {
		log.Warn("Disable \"%s\" for level \"%s\"", tag, fc.Level)
		return
	}

	var appender l4g.Appender
	switch tag {
	case "color":
		appender = colorlog.NewAppender()
	case "file":
		appender = filelog.NewAppender("_test.log", 0)
	case "socket":
		appender = socketlog.NewAppender("udp", "127.0.0.1:12124")
	case "xml":
		appender = filelog.NewAppender("_test.log", 0)
		appender.SetOption("head","<log created=\"%D %T\">%R")
		
		appender.SetOption("pattern", 
`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>%R`)
		
		appender.SetOption("foot", "</log>%R")
	default:
		log.Error("Unknown appender <%s>", tag)
		return
	}

	ok := true
	for _, prop := range fc.Props {
		err := appender.SetOption(prop.Name, strings.Trim(prop.Value, " \r\n"))
		if err != nil {
			log.Error(err)
			ok = false
		}
	}
	if !ok {
		return
	}
 	
	filters.Add(tag, lvl, appender)
}

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

	xc := new(LoggerConfig)
	if err := xml.Unmarshal(buf, xc); err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse XML configuration. %s\n", err)
		os.Exit(1)
	}
	log.Debug("Total configuration: %d", len(xc.Filters))
	// fmt.Println(lc)

	for _, fc := range xc.Filters {
		loadFilter(&fc)
	}
	log.Debug("Total appenders installed: %d", len(*filters))
	if filt := filters.Get("color"); filt != nil {
		// disable default console writer
		log.SetOutput(nil)
	}
	log.SetFilters(filters)

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Trace("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	log.SetFilters(nil)
	filters.Close()

	PrintFile("_test.log")
	os.Remove("_test.log")
	PrintFile("_trace.xml")
	os.Remove("_trace.xml")
}

