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
	filename = "_flw.log"
	oldfiles = "_flw.*.log"
)

func CheckTimer(cycle int64, delay0 int64) {
	fmt.Println("cycle:", cycle, "delay0:", delay0)
	nrt := time.Now()
	if delay0 < 0 { // Now + cycle
		nrt = nrt.Add(time.Duration(cycle) * time.Second)
	} else { // tomorrow midnight (Clock 0) + delay0
		tomorrow := nrt.Add(24 * time.Hour)
        nrt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 
						0, 0, 0, 0, tomorrow.Location())
		nrt = nrt.Add(time.Duration(delay0) * time.Second)
	}
	fmt.Println("nrt:", nrt, "now:", time.Now())
	fmt.Println("First timer:", nrt.Sub(time.Now()))
}

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
	fmt.Println("Every 10 minutes")
	CheckTimer(600, -1)
	fmt.Println("---\nEvery midnight")
	CheckTimer(86400, 0)
	fmt.Println("---\nEvery 3:00am")
	CheckTimer(86400, 10800)
	fmt.Println("---\nEvery weekly midnight")
	CheckTimer(86400 * 7, 0)

	// Get a new logger instance
	log := l4g.New(l4g.FINE)

	// Create a default logger that is logging messages of FINE or higher
	log.AddFilter("file", l4g.FINE, filelog.NewFileLogWriter(filename, 0))
	log.Finest("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Critical("Time to close out!")
	log.CloseFilters()

	PrintFile(filename)
	// Remove the file so it's not lying around
	err := os.Remove(filename)
	if err != nil {
		fmt.Println(err)
	}

	/* Can also specify manually via the following: (these are the defaults) */
	flw := filelog.NewFileLogWriter(filename, 10)
	flw.Set("cycle", 5)
	flw.Set("delay0", -1)
	flw.Set("format", "[%D %T] [%L] (%x) %M")
	flw.Set("maxsize", "5k")
	log.AddFilter("file", l4g.FINE, flw)

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		time.Sleep(1 * time.Second)
		for i := 0; i < 200 / (j+1); i++ {
			log.Finest("Everything is created now (notice that I will not be printing to the file)")
			log.Info("%d. The time is now: %s", j, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Critical("Time to close out!")
		}
		time.Sleep(4 * time.Second)
	}
	// Close the log filters
	log.CloseFilters()

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
