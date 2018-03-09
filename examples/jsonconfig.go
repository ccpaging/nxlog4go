package main

import (
	"encoding/json"
	"flag"
	"os"
	"fmt"
	"bufio"
	"io"
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

// Load appenders' configuration 
func ConfigureLogger(logger *l4g.Logger, contents []byte, debug bool) {
	if debug {
		logger.SetLevel(l4g.DEBUG)
	} else {
		logger.SetLevel(l4g.INFO)
	}

 	// fmt.Println(string(contents))
	// decode json logger config
	lc := new(l4g.LoggerConfig)
	if err := json.Unmarshal(contents, lc); err != nil {
		Error("ConfigureLogger", "Can't parse %v %v", contents, err)
		return
	}

	if len(lc.Filters) <= 0 {
		Warn("ConfigureLogger", "Filters is NIL. The struct name should be 'filters'")
		return
	}
 	
	Info("ConfigureLogger", "Total configuration: %d", len(lc.Filters))
	// fmt.Println(lc)
	fs := l4g.NewFilters()
	// Preload appenders which may be in configuration
	fs.Preload("color", colorlog.NewAppender())
	fs.Preload("file", filelog.NewAppender("_test.log", 0))
	fs.Preload("socket", socketlog.NewAppender("udp", "127.0.0.1:12124"))

	fs.LoadConfiguration(lc.Filters)
	if filt, isExist := (*fs)["color"]; isExist {
		if debug {
			filt.Level = l4g.DEBUG
		}
		// New console appender loaded and disable default writer
		logger.SetOutput(nil)
	}
	logger.SetFilters(fs)
	Info("ConfigureLogger", "Total appenders installed: %d", len(*fs))
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
	ConfigureLogger(log, c.LogConfig, *debug)

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	// Close all appenders in logger
	log.Shutdown()

	PrintFile("_test.log")
	os.Remove("_test.log")
}

