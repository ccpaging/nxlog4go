// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"testing"
)

func TestColorFormatting(t *testing.T) {
	var colorBytes = [][]byte{
		Red.Wrap([]byte("Red Color")), LightRed.Wrap([]byte("Light Red")),
		Green.Wrap([]byte("Green Color")), LightGreen.Wrap([]byte("Light Green")),
		Yellow.Wrap([]byte("Yellow Color")), LightYellow.Wrap([]byte("Light Yellow")),
		Blue.Wrap([]byte("Blue Color")), LightBlue.Wrap([]byte("Light Blue")),
		Magenta.Wrap([]byte("Magenta Color")), LightMagenta.Wrap([]byte("Light Magenta")),
		Cyan.Wrap([]byte("Cyan Color")), LightCyan.Wrap([]byte("Light Cyan")),
		White.Wrap([]byte("White Color")), LightWhite.Wrap([]byte("Light White")),
		Gray.Wrap([]byte("Gray Color")),
	}
	for i := 0; i < len(colorBytes); i++ {
		print(string(colorBytes[i]))
		i++
		if i < len(colorBytes) {
			println(" -", string(colorBytes[i]))
		} else {
			println("")
		}
	}
}
