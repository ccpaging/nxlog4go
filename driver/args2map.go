// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package driver

import (
	"fmt"
)

// ArgsToMap turn args []interface{} to map dictionary and strings array index.
// Return: map[string]interface{}, []string, error
func ArgsToMap(args []interface{}) (map[string]interface{}, []string, error) {
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
