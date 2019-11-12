// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

var (
	// Normal colors
	nBlack   = []byte{'\033', '[', '3', '0', 'm'}
	nRed     = []byte{'\033', '[', '3', '1', 'm'}
	nGreen   = []byte{'\033', '[', '3', '2', 'm'}
	nYellow  = []byte{'\033', '[', '3', '3', 'm'}
	nBlue    = []byte{'\033', '[', '3', '4', 'm'}
	nMagenta = []byte{'\033', '[', '3', '5', 'm'}
	nCyan    = []byte{'\033', '[', '3', '6', 'm'}
	nWhite   = []byte{'\033', '[', '3', '7', 'm'}
	// Bright colors
	bBlack   = []byte{'\033', '[', '3', '0', ';', '1', 'm'}
	bRed     = []byte{'\033', '[', '3', '1', ';', '1', 'm'}
	bGreen   = []byte{'\033', '[', '3', '2', ';', '1', 'm'}
	bYellow  = []byte{'\033', '[', '3', '3', ';', '1', 'm'}
	bBlue    = []byte{'\033', '[', '3', '4', ';', '1', 'm'}
	bMagenta = []byte{'\033', '[', '3', '5', ';', '1', 'm'}
	bCyan    = []byte{'\033', '[', '3', '6', ';', '1', 'm'}
	bWhite   = []byte{'\033', '[', '3', '7', ';', '1', 'm'}

	reset = []byte{'\033', '[', '0', 'm'}
)

func setColor(e *Entry) bool {
	if out := e.logger.out; out != nil {
		switch e.Level {
		case CRITICAL:
			out.Write(bRed)
		case ERROR:
			out.Write(nRed)
		case WARNING:
			out.Write(bYellow)
		case INFO:
		case TRACE:
			out.Write(nMagenta)
		case DEBUG:
			out.Write(nGreen)
		case FINE:
			out.Write(nCyan)
		case FINEST:
			out.Write(nBlue)
		default:
		}
	}
	return true
}

func resetColor(e *Entry, n int, err error) {
	if out := e.logger.out; out != nil {
		out.Write(reset)
	}
}
