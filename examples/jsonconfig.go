package main

import (
	"encoding/json"
	"os"
	"fmt"
	"bufio"
	"io"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/file"
	"github.com/ccpaging/nxlog4go/socket"
)

var filename string = "config.json"

var log = l4g.GetLogger()

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
	fd, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Can't load json config file: %s %v", filename, err))
	}
	defer fd.Close()

	// decode json
	type Config struct {
		LogConfig json.RawMessage
	}

	c := Config{}
	err = json.NewDecoder(fd).Decode(&c)
	if err != nil {
		panic(fmt.Sprintf("Can't parse json config file: %s %v", filename, err))
	}
	
	// decode json logger config
	lc := new(l4g.LoggerConfig)
	if err := json.Unmarshal(c.LogConfig, lc); err != nil {
		panic(fmt.Sprintf("Can't parse LogConfig: %v %v", c.LogConfig, err))
	}

	if len(lc.Filters) <= 0 {
		panic(fmt.Sprintf("<filters> section is nil.\n%s", c.LogConfig))
	}

	fs := l4g.NewFilters()

	// Preload appender may used in configuration
	fs.Preload("color", colorlog.NewAppender())
	fs.Preload("file", filelog.NewAppender("_test.log", 0))
	fs.Preload("socket", socketlog.NewAppender("udp", "127.0.0.1:12124"))
	fmt.Println(len(*fs), "appenders pre-installed")

	fs.LoadConfiguration(lc.Filters)
	if len(*fs) > 0 {
		log.SetFilters(fs)
		// disable default console writer
		log.SetOutput(nil)
	}
	fmt.Println(len(*fs), "appenders configured ok")

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	log.SetFilters(nil)
	fs.Close()

	PrintFile("_test.log")
	os.Remove("_test.log")
}

