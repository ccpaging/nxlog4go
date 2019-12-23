package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/console"
	"github.com/ccpaging/nxlog4go/driver"
)

var isColor = (os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb") || os.Getenv("ConEmuANSI") == "ON"

func main() {
	fmt.Println("    TERM:", os.Getenv("TERM"))
	fmt.Println("    ConEmuANSI:", os.Getenv("ConEmuANSI"))
	// Get global logger of logs
	// disable default term output of logger
	logger := log.GetLogger().SetOutput(ioutil.Discard)
	a, _ := driver.Open("console", "",
		"level", "FINEST",
		"color", true)

	// fs := log.NewFilters().Add("console", log.FINEST, a)
	f := log.NewFilter(0, nil, a)
	logger.Attach(f)

	log.Finest("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Fine("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Trace("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	logger.Detach(f)
	f.Close()
}
