package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
	"path/filepath"
	l4g "github.com/ccpaging/nxlog4go"
	"github.com/ccpaging/nxlog4go/file"
)

const (
	filename = "_fa.log"
	oldfiles = "_fa.*.log"
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

func init() {
	// Enable internal logger
	l4g.GetLogLog().SetLevel(l4g.TRACE)
}

func main() {
	// Get a new logger instance
	log := l4g.New(l4g.FINE)

	// Create a default logger that is logging messages of FINE or higher
	fs0 := l4g.NewFilters().Add("file", l4g.FINE, filelog.NewAppender(filename, 0))
	log.SetFilters(fs0)
	log.Finest("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Critical("Time to close out!")
	log.SetFilters(nil)
	fs0.Close()

	PrintFile(filename)
	// Remove the file so it's not lying around
	err := os.Remove(filename)
	if err != nil {
		fmt.Println(err)
	}

	/* Can also specify manually via the following: (these are the defaults) */
	fa := filelog.NewAppender(filename, 10)
	fa.Set("format", "[%D %T] [%L] (%x) %M")
	fa.Set("cycle", 5)
	fa.Set("delay0", -1)
	fa.Set("maxsize", "5k")
	fs1 := l4g.NewFilters().Add("file", l4g.FINE, fa)
	log.SetFilters(fs1)

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		time.Sleep(1 * time.Second)
		for i := 0; i < 200 / (j+1); i++ {
			log.Finest("Everything is created now (notice that I will not be printing to the file)")
			log.Info("%d. The time is now: %s", j, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Critical("Time to close out!")
		}
		//time.Sleep(1 * time.Second)
	}
	// Close the log filters
	log.SetFilters(nil)
	fs1.Close()

	PrintFile(filename)
	os.Remove(filename)

	// contains a list of all files in the current directory
	files, _ := filepath.Glob(oldfiles)
    fmt.Printf("%d files match %s\n", len(files), oldfiles)
    for _, f := range files {
		fmt.Printf("Remove %s\n", f)
		os.Remove(f)
    }
}
