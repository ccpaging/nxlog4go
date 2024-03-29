// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package driver

import (
	"fmt"
	"strings"
)

// ArgsToMap turn args...interface{} to map dictionary and strings array index.
// Return: map[string]interface{}, []string, error
func ArgsToMap(args ...interface{}) (map[string]interface{}, []string, error) {
	d := make(map[string]interface{}, len(args)/2)
	var o []string
	for i := 0; i < len(args); i += 2 {
		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			return d, o, fmt.Errorf("the number of arguments should be odd but %d", len(args))
		}
		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		s, ok := key.(string)
		if !ok {
			// Subsequent errors are likely, so allocate once up front.
			return d, o, fmt.Errorf("the key %#v of type %T at %d should be string", key, key, i)
		}

		o = append(o, s)
		switch v := val.(type) {
		case string:
			d[s] = val.(string)
		case error:
			d[s] = v.Error()
		case func() string:
			d[s] = v()
		default:
			d[s] = val
		}
	}
	return d, o, nil
}

// LazyArgsToMap turn args...interface{} to map dictionary and strings array index.
// Return: map[string]interface{}, []string
func LazyArgsToMap(args ...interface{}) (map[string]interface{}, []string) {
	out := make(map[string]interface{}, len(args)/2)
	var index []string

	for i, j := 0, 0; i < len(args); i++ {
		value := args[i]
		k, ok := value.(string)
		if !ok {
			// Not string
			k = fmt.Sprintf("key%d", j)
			j++
		} else {
			i++
		}
		if i == len(args) {
			// this element is a dangling key.
			value = "nil"
		} else {
			value = args[i]
		}

		switch v := value.(type) {
		case string:
			out[k] = v
		case error:
			out[k] = v.Error()
		case func() string:
			out[k] = v()
		default:
			out[k] = value
		}
		index = append(index, k)
	}
	return out, index
}

// ArgsToValues turn args []interface{} to interface{} array.
// Return: []interface{}
func ArgsToValues(args ...interface{}) []interface{} {
	var o []interface{}
	for _, val := range args {
		switch v := val.(type) {
		case string:
			o = append(o, val.(string))
		case error:
			o = append(o, v.Error())
		case func() string:
			o = append(o, v())
		default:
			o = append(o, val)
		}
	}
	return o
}

// ArgsToString builds a format string by the arguments
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
