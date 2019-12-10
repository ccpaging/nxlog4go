// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package cast

import (
	"fmt"
	"strconv"
	"time"
)

// ToString casts an interface to a string type.
// Default: ""
func ToString(i interface{}) (s string, err error) {
	s = ""
	err = nil

	switch i.(type) {
	case string:
		s = i.(string)
	case []byte:
		s = string(i.([]byte))
	default:
		err = fmt.Errorf("unable to cast %#v of type %T to String", i, i)
	}
	return
}

// ToBool casts an interface to a bool type.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// Any other value returns an error.
// Default: false
//
// See also: http://golang.org/pkg/strconv/#ParseBool
func ToBool(i interface{}) (b bool, err error) {
	b = false
	err = nil

	switch i.(type) {
	case bool:
		b = i.(bool)
	case int:
		if i.(int) > 0 {
			b = true
		}
	case string:
		return strconv.ParseBool(i.(string))
	default:
		err = fmt.Errorf("unable to cast %#v of type %T to Bool", i, i)
	}
	return
}

func strToNumSuffix(s string, base int64) (int64, error) {
	var multi int64 = 1
	if len(s) > 1 {
		switch s[len(s)-1] {
		case 'G', 'g':
			multi *= base
			fallthrough
		case 'M', 'm':
			multi *= base
			fallthrough
		case 'K', 'k':
			multi *= base
			s = s[0 : len(s)-1]
		}
	}
	n, err := strconv.ParseInt(s, 0, 0)
	return n * multi, err
}

// ToInt casts an interface to an int type.
// Parse a string with K/M/G suffixes based on thousands (1000) or 2^10 (1024)
// Default: 0
//
// See also: http://golang.org/pkg/strconv/#ParseInt
func ToInt(i interface{}) (n int, err error) {
	n = 0
	err = nil

	switch i.(type) {
	case int:
		n = i.(int)
	case int64:
		n = int(i.(int64))
	case string:
		var i64 int64
		i64, err = strToNumSuffix(i.(string), 1024)
		n = int(i64)
	default:
		err = fmt.Errorf("unable to cast %#v of type %T to Int", i, i)
	}
	return
}

// ToInt64 casts an interface to an int64 type.
// Parse a string with K/M/G suffixes based on thousands (1000) or 2^10 (1024)
// Default: 0
//
// See also: http://golang.org/pkg/strconv/#ParseInt
func ToInt64(i interface{}) (n int64, err error) {
	n = 0
	err = nil

	switch i.(type) {
	case int:
		v := i.(int)
		n = int64(v)
	case int64:
		n = i.(int64)
	case string:
		n, err = strToNumSuffix(i.(string), 1024)
	default:
		err = fmt.Errorf("unable to cast %#v of type %T to Int64", i, i)
	}
	return
}

// ToSeconds casts an interface to an seconds.
// Parse a string with time.ParseDuration. Valid time units are:
//  "ns", "us", "ms", "s", "m", "h"
// Default: 0
func ToSeconds(i interface{}) (n int64, err error) {
	n = 0
	err = nil

	switch i.(type) {
	case int:
		n = int64(i.(int))
	case int64:
		n = i.(int64)
	case string:
		var dur time.Duration
		dur, err = time.ParseDuration(i.(string))
		n = int64(dur) / int64(time.Second)
	default:
		err = fmt.Errorf("unable to cast %#v of type %T to Seconds", i, i)
	}
	return
}
