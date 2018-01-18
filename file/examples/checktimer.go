package main

import (
	"fmt"
	"time"
)

const (
	filename = "_flw.log"
	oldfiles = "_flw.*.log"
)

func CheckTimer(cycle, clock int64) {
	if cycle <= 0 {
		cycle = 86400
	}
	fmt.Println("cycle:", cycle, "clock:", clock)
	nrt := time.Now()
	if clock < 0 { // Now + cycle
		nrt = nrt.Add(time.Duration(cycle) * time.Second)
	} else { // tomorrow midnight (Clock 0) + delay0
		tomorrow := nrt.Add(time.Duration(cycle) * time.Second)
        nrt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 
						0, 0, 0, 0, tomorrow.Location())
		nrt = nrt.Add(time.Duration(clock) * time.Second)
	}
	fmt.Println("nrt:", nrt, "now:", time.Now())
	fmt.Println("First timer:", nrt.Sub(time.Now()))
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
}
