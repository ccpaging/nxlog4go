package main

import (
	"encoding/json"
	"os"
	"fmt"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/file"
	"github.com/ccpaging/nxlog4go/socket"
)

var filename string = "config.json"

var log = l4g.GetLogger()

func main() {
	// disable default console writer
	log.SetOutput(nil)

	// Open config file
	fd, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Can't load json config file: %s %v", filename, err))
	}
	defer fd.Close()

	// decode json
	type Config struct {
		LoggerConfig json.RawMessage
	}

	c := Config{}
	err = json.NewDecoder(fd).Decode(&c)
	if err != nil {
		panic(fmt.Sprintf("Can't parse json config file: %s %v", filename, err))
	}
	
	// decode json logger config
	lc := new(l4g.LoggerConfig)
	if err := json.Unmarshal(c.LoggerConfig, lc); err != nil {
		panic(fmt.Sprintf("Can't parse json config file: %s %v", filename, err))
	}

	var	lw l4g.LogWriter
	for _, fc := range lc.Filters {
		bad, enabled, lvl := l4g.CheckFilterConfig(fc)
	
		if bad {
			os.Exit(1)
		}
	
		switch fc.Type {
		case "console":
			lw = NewColorLogWriter()
		case "file":
			lw = NewFileLogWriter(DefaultFileName, 0)
		case "socket":
			lw = NewSocketLogWriter(DefaultSockProto, DefaultSockEndPoint)
		default:
			panic(fmt.Sprintf("Unknown filter type \"%s\"", fc.Type))
		}
	
		if lw == nil {
			panic(fmt.Sprintf("Unknown filter type \"%s\"", fc.Type))
			fmt.Fprintf(os.Stderr, "LoadConfiguration: LogWriter is nil. %v\n", fc)
			os.Exit(1)
		}
	
		log.AddFilter(fc.Tag, lvl, lw)
	}

	// And now we're ready!
	log.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	log.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	log.Info("About that time, eh chaps?")

	log.CloseFilters()

	os.Remove("_test.log")
}

