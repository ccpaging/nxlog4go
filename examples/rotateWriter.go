package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
)

const (
	filename = "_rfw.log"
	backups  = "_rfw.*"
)

func main() {
	// Enable internal logger
	l4g.GetLogLog().Set("level", l4g.TRACE)

	// Can also specify manually via the following: (these are the defaults)
	rfw := l4g.NewRotateFileWriter(filename, true).Set("maxsize", 1024*5).Set("maxbackup", 10)
	ww := io.MultiWriter(os.Stderr, rfw)
	// Get a new logger instance
	log := l4g.NewLogger(l4g.FINEST).SetOutput(ww).Set("pattern", "[%D %T] [%L] (%s) %M\n")

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		for i := 0; i < 200/(j+1); i++ {
			log.Finest("Everything is created now (notice that I will not be printing to the file)")
			log.Info("%d. The time is now: %s", j, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Critical("Time to close out!")
		}
	}

	rfw.Close()
	fmt.Printf("Remove %s\n", filename)
	os.Remove(filename)

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(backups)
	fmt.Printf("%d files match %s\n", len(files), backups)
	for _, f := range files {
		fmt.Printf("Remove %s\n", f)
		os.Remove(f)
	}
}
