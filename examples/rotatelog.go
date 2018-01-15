package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
	"path/filepath"

	l4g "github.com/ccpaging/nxlog4go"
)

const (
	filename = "_fw.log"
	backups = "_fw.*"
)

// Print what was logged to the file (yes, I know I'm skipping error checking)
func PrintFile(fn string) {
	fd, _ := os.Open(fn)
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
	// Can also specify manually via the following: (these are the defaults)
	frw := l4g.NewRotateFileWriter(filename).SetMaxSize(1024 * 5).SetMaxBackup(10)
	ww := io.MultiWriter(os.Stderr, frw)
	// Get a new logger instance
	log := l4g.New(l4g.FINEST).SetOutput(ww).SetFormat("[%D %T] [%L] (%s) %M")

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		for i := 0; i < 200 / (j+1); i++ {
			log.Finest("Everything is created now (notice that I will not be printing to the file)")
			log.Info("%d. The time is now: %s", j, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Critical("Time to close out!")
		}
	}

	frw.Close()
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
