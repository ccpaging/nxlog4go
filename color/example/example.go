package main

import (
	"time"
	"os"
	"fmt"

	log "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/color"
)

func main() {
	fmt.Println("    TERM:", os.Getenv("TERM"))
	fmt.Println("    ConEmuANSI:", os.Getenv("ConEmuANSI"))  

	// Get global logger of logs
	logger := log.GetLogger()
	// disable default term output of logger
	logger.SetOutput(nil)
	fs := log.NewFilters().Add("color", log.FINEST, colorlog.NewLogWriter())
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
