package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/driver"
	_ "github.com/ccpaging/nxlog4go/file"
)

var logFile = "_test.xml"
var removeFiles = "_test*.xml"

func main() {
	// Enable internal logger
	log.GetLogLog().SetOptions("level", log.TRACE, "caller", true, "format", "[%D %T] [%L] (%S:%N) \t%M")
	defer log.GetLogLog().SetOptions("level", log.CRITICAL+1)

	logger := log.GetLogger().SetOptions("level", log.ERROR)

	a, err := driver.Open("xml", logFile,
		"level", log.FINE,
		"rotate", 1,
		"cycle", "-1s",
		"maxlines", "100")
	if err != nil {
		fmt.Println(err)
		return
	}

	logger.Attach(log.NewFilter(log.FINE, nil, a))

	for i := 0; i < 125; i++ {
		log.Info("%d: The time is now: %s", i+1, time.Now().Format("15:04:05 MST 2006/01/02"))
	}

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
