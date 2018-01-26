package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
)

const (
	filename = "_fw.log"
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
	fbw := l4g.NewFileBufWriter(filename)
	ww := io.MultiWriter(os.Stderr, fbw)
	// Get a new logger instance
	log := l4g.New(l4g.FINEST).SetOutput(ww)

	log.Finest("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Critical("Time to close out!")
	fbw.Close()

	PrintFile(filename)
	// Remove the file so it's not lying around
	os.Remove(filename)
}
