package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/file"
)

var logFile = "_test.json"
var removeFiles = "_test*.json"

func main() {
	// Enable internal logger
	log.GetLogLog().Set("level", log.TRACE, "caller", true, "format", "[%D %T] [%L] (%S:%N) \t%M")
	defer log.GetLogLog().Set("level", log.CRITICAL+1)

	logger := log.GetLogger().Set("level", log.ERROR)

	a, err := log.Open("json", logFile,
		"level", log.TRACE,
		"rotate", 0,
		"cycle", "5s",
		"maxsize", "5k")
	if err != nil {
		fmt.Println(err)
		return
	}

	logger.Attach(log.NewFilter(log.TRACE, nil, a))

	for i := 0; i < 25; i++ {
		for j := 0; j < 5; j++ {
			log.Info("%d: The time is now: %s", i, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		}
		time.Sleep(1 * time.Second)
	}

	log.Finest("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Fine("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Trace("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	logger.Set("utc", true) // console only
	a.Set("utc", true)      // file only
	log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	logger.Close()

	if contents, err := ioutil.ReadFile(logFile); err == nil {
		fmt.Println(string(contents))
	} else {
		fmt.Println(err)
	}

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(removeFiles)
	fmt.Printf("%d files match %s\n", len(files), removeFiles)
	for _, fname := range files {
		fmt.Printf("Remove %q\n", fname)
		err := os.Remove(fname)
		if err != nil {
			fmt.Println(err)
		}
	}
}
