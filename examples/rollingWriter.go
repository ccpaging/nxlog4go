package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/rolling"
)

const (
	filename = "_rfw.log"
	backups  = "_rfw*.log"
)

func main() {
	// Enable internal logger
	l4g.GetLogLog().Set("level", l4g.TRACE)

	// Can also specify manually via the following: (these are the defaults)
	rfw := rolling.NewWriter(filename, 1024*5)
	ww := io.MultiWriter(os.Stderr, rfw)
	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINEST).SetOutput(ww).Set("format", "[%D %T] [%L] (%S) %M")

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		for i := 0; i < 200/(j+1); i++ {
			log.Finest("Everything is created now (notice that I will not be printing to the file)")
			log.Info("%d. The time is now: %s", j, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Critical("Time to close out!")
		}
	}

	rfw.Close()

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(backups)
	fmt.Printf("%d files match %s\n", len(files), backups)
	for _, f := range files {
		fmt.Printf("Remove %s\n", f)
		os.Remove(f)
	}
}
