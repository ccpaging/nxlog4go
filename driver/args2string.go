// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package driver

import (
	"fmt"
	"strings"
)

// AgrsToString builds a format string by the arguments
// Return a format string which depends on the first argument:
//
// - arg[0] is a string
//   When given a string as the first argument, this behaves like fmt.Sprintf
//   the first argument is interpreted as a format for the latter arguments.
//
// - arg[0] is a func()string
//   When given a closure of type func()string, this logs the string returned by
//   the closure if it will be logged.  The closure runs at most one time.
//
// - arg[0] is interface{}
//   When given anything else, the return message will be each of the arguments
//   formatted with %v and separated by spaces (ala Sprint).
func ArgsToString(arg0 interface{}, args ...interface{}) (s string) {
	switch first := arg0.(type) {
	case string:
		if len(args) == 0 {
			s = first
		} else {
			// Use the string as a format string
			s = fmt.Sprintf(first, args...)
		}
	case func() string:
		// Log the closure (no other arguments used)
		s = first()
	default:
		// Build a format string so that it will be similar to Sprint
		if len(args) == 0 {
			s = fmt.Sprint(first)
		} else {
			s = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		}
	}
	return
}
