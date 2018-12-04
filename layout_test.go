// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package nxlog4go

import (
	"testing"
	"time"
	"bytes"
)

func TestFormatHMS(t *testing.T) {
	now := time.Now()
	//year, month, day := now.Date()
	out := bytes.NewBuffer(make([]byte, 0, 64))
	formatHMS(out, &now, ':')
	s0 := string(out.Bytes())
	s1 := now.Format("15:04:05")
	if s0 != s1 {
		t.Errorf("Incorrect time format: %s should be %s", s0, s1)
	}
}

func TestFormatDMY(t *testing.T) {
	now := time.Now()
	out := bytes.NewBuffer(make([]byte, 0, 64))
	formatDMY(out, &now, '/')
	s0 := string(out.Bytes())
	s1 := now.Format("02/01/06")
	if s0 != s1 {
		t.Errorf("Incorrect time format: %s should be %s", s0, s1)
	}
}

func TestFormatCYMD(t *testing.T) {
	now := time.Now()
	out := bytes.NewBuffer(make([]byte, 0, 64))
	formatCYMD(out, &now, '/')
	s0 := string(out.Bytes())
	s1 := now.Format("2006/01/02")
	if s0 != s1 {
		t.Errorf("Incorrect time format: %s should be %s", s0, s1)
	}
}
