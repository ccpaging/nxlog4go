package main

import (
	"sync"
	"time"

	l4g "github.com/ccpaging/nxlog4go"
)

var (
	log1 = l4g.NewLogger(l4g.DEBUG).SetOptions("prefix", "expl", "format", "[%P] %T %D %Z] [%L] (%S:%N) %M")
	log2 = log1.Clone().SetOptions("prefix", "exp2")
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		log1.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log1.Set("caller", false)
		log1.Trace("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log1.Layout().Set("utc", true)
		log1.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log1.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log1.Error("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log1.Critical("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		wg.Done()
	}()

	go func() {
		log2.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log2.Trace("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log2.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log2.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log2.Error("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		log2.Critical("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
		wg.Done()
	}()

	wg.Wait()
}
