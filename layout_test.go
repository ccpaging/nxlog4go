// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package nxlog4go

import (
	"testing"
	"time"
)

func TestFormatHHMMSS(t *testing.T) {
	now := time.Now()
	//year, month, day := now.Date()
	hour, minute, second := now.Clock()
	var b []byte
	formatHHMMSS(&b, hour, minute, second)
	s0 := string(b)
	s1 := now.Format("15:04:05")
	if s0 != s1 {
		t.Errorf("Incorrect time format: %s should be %s", s0, s1)
	}
}

func TestFormatCCYYMMDD(t *testing.T) {
	now := time.Now()
	year, month, day := now.Date()
	var b []byte
	formatCCYYMMDD(&b, year / 100, year % 100, int(month), int(day), '/')
	s0 := string(b)
	s1 := now.Format("2006/01/02")
	if s0 != s1 {
		t.Errorf("Incorrect time format: %s should be %s", s0, s1)
	}
}

func TestFormatDDMMYY(t *testing.T) {
	now := time.Now()
	year, month, day := now.Date()
	var b []byte
	formatDDMMYY(&b, year % 100, int(month), int(day))
	s0 := string(b)
	s1 := now.Format("02/01/06")
	if s0 != s1 {
		t.Errorf("Incorrect time format: %s should be %s", s0, s1)
	}
}
