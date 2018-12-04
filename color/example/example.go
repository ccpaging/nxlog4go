package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
)

func main() {
	fmt.Println("    TERM:", os.Getenv("TERM"))
	fmt.Println("    ConEmuANSI:", os.Getenv("ConEmuANSI"))
	// Get global logger of logs
	// disable default term output of logger
	logger := log.GetLogger().SetOutput(ioutil.Discard)
	fs := log.NewFilters().Add("color", log.FINEST, colorlog.NewColorAppender(os.Stderr))
	logger.SetFilters(fs)
	log.Finest("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Fine("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Trace("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	logger.SetFilters(nil)
	fs.Close()
}
