// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package file

import (
	"testing"
	"time"
)

var now = time.Unix(0, 1234567890123456789).In(time.UTC)

func nextTime(now time.Time, cycle, clock int64) time.Time {
	if cycle < 5 {
		cycle = 5
	}

	if cycle < dayToSecs {
		// Now + cycle
		return now.Add(time.Duration(cycle) * time.Second)
	}
	// now + cycle
	t := now.Add(time.Duration(cycle) * time.Second)
	// back to midnight
	t = time.Date(t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0, t.Location())
	// midnight + cycle % 86400
	t = t.Add(time.Duration(clock) * time.Second)
	return t
}

func TestNextTime(t *testing.T) {
	d0, d1 := nextTime(now, 600, -1).Sub(now), time.Duration(10*time.Minute)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (10 minutes): %v should be %v", d0, d1)
	}
	// Correct invalid value cycle = 300，clock = 0 to clock = -1
	// for cycle < 86400
	d0, d1 = nextTime(now, 300, 0).Sub(now), time.Duration(5*time.Minute)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (5 minutes): %v should be %v", d0, d1)
	}

	t1 := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	d0, d1 = nextTime(now, 86400, 0).Sub(now), t1.Sub(now)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (next midnight): %v should be %v", d0, d1)
	}

	t1 = time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
	d0, d1 = nextTime(now, 86400, 10800).Sub(now), t1.Sub(now)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (next 3:00am): %v should be %v", d0, d1)
	}

	t1 = time.Date(now.Year(), now.Month(), now.Day()+7, 0, 0, 0, 0, now.Location())
	d0, d1 = nextTime(now, 86400*7, 0).Sub(now), t1.Sub(now)
	if d0 != d1 {
		t.Errorf("Incorrect nextTime duration (next weekly midnight): %v should be %v", d0, d1)
	}
}
