package main

import (
	"encoding/json"
	"flag"
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/file"
	"github.com/ccpaging/nxlog4go/socket"
)

var (
	debug    = flag.Bool("debug", false, "")
	filename = flag.String("conf", "config.json", "config file")
)

var log = l4g.GetLogger().SetCaller(false).SetPattern("[%T] [%L] (%s) %M\n")
var filters = l4g.NewFilters()

// Wrapper for app developing
func Debug(source string, arg0 interface{}, args ...interface{}) {
	log.Log(l4g.DEBUG, source, arg0, args ...)
}

func Info(source string, arg0 interface{}, args ...interface{}) {
	log.Log(l4g.INFO, source, arg0, args ...)
}

func Warn(source string, arg0 interface{}, args ...interface{}) {
	log.Log(l4g.WARNING, source, arg0, args ...)
}

func Error(source string, arg0 interface{}, args ...interface{}) {
	log.Log(l4g.ERROR, source, arg0, args ...)
}

// Appender's properties
type AppenderProp struct {
	Name  string
	Value string
}

type FilterConfig struct {
	Tag		 string
	Level    string
	Props	 []AppenderProp `json:"properties"`
}

type LoggerConfig struct {
	Filters  []FilterConfig `json:"filters"`
}

func loadFilter(fc *FilterConfig) {
	if len(fc.Tag) == 0 {
		Error("config", "Required child tag")
		return
	}
	tag := strings.ToLower(fc.Tag)
	if len(fc.Level) == 0 {
		Error("config", "Required child level")
		return
	}
	lvl := l4g.GetLevel(fc.Level)
	if lvl >= l4g.SILENT {
		Warn("config", "Disable \"%s\" for level \"%s\"", tag, fc.Level)
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
	default:
		Error("config", "Unknown appender <%s>", tag)
		return
	}

	ok := true
	for _, prop := range fc.Props {
		err := appender.SetOption(prop.Name, strings.Trim(prop.Value, " \r\n"))
		if err != nil {
			Error("config", err)
			ok = false
		}
	}
	if !ok {
		return
	}
 	
	filters.Add(tag, lvl, appender)
}


// Load appenders' configuration 
func loadConfiguration(contents []byte, debug bool) {
	if debug {
		log.SetLevel(l4g.DEBUG)
	} else {
		log.SetLevel(l4g.INFO)
	}

 	// fmt.Println(string(contents))
	// decode json logger config
	lc := new(LoggerConfig)
	if err := json.Unmarshal(contents, lc); err != nil {
		Error("config", "Can't parse %v %v", contents, err)
		return
	}

	if len(lc.Filters) <= 0 {
		Warn("config", "Filters is NIL. The struct name should be 'filters'")
		return
	}
 	
	Info("config", "Total configuration: %d", len(lc.Filters))
	// fmt.Println(lc)

	for _, fc := range lc.Filters {
		loadFilter(&fc)
	}

	if filt := filters.Get("color"); filt != nil {
		if debug {
			filt.Level = l4g.DEBUG
		}
		// New console appender loaded and disable default writer
		log.SetOutput(nil)
	}
	log.SetFilters(filters)
	Info("config", "Total appenders installed: %d", len(*filters))
	// Close filters by call log.Shutdown() when program exit
}

// Print what was logged to the file (yes, I know I'm skipping error checking)
func PrintFile(fn string) {
	fd, err := os.Open(fn)
	if err != nil {
		fmt.Println(err)
		return
	}
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
	flag.Parse()

	// Open config file
	fd, err := os.Open(*filename)
	if err != nil {
		panic(fmt.Sprintf("Can't load json config file: %s %v", filename, err))
	}

	type Config struct {
		LogConfig json.RawMessage
	}
	c := Config{}

	// decode json
	err = json.NewDecoder(fd).Decode(&c)
	fd.Close()
	if err != nil {
		panic(fmt.Sprintf("Can't parse json config file: %s %v", filename, err))
	}

	// Configure logger
	loadConfiguration(c.LogConfig, *debug)

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	// Close all appenders in logger
	log.Shutdown()

	PrintFile("_test.log")
	os.Remove("_test.log")
}

