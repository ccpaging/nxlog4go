// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package patt

// NewCSVLayout creates a new layout encoding log Recorder as CSV format.
func NewCSVLayout(args ...interface{}) *PatternLayout {
	csvFormat := "%D|%T|%L|%P|%S:%N|%M%F"
	lo := NewLayout(csvFormat, args...)
	lo.SetOptions("fieldsEncoder", "csv")
	return lo
}
