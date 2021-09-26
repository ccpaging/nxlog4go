// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package patt

// NewJSONLayout creates a new layout encoding log Recorder as JSON format.
func NewJSONLayout(args ...interface{}) *PatternLayout {
	return NewJSONFieldsLayout(args...)
}

// NewJSONFieldsLayout creates a new layout encoding log Recorder as JSON format with Feilds map[string]interface{}.
func NewJSONFieldsLayout(args ...interface{}) *PatternLayout {
	jsonFormat := "{\"Level\":%l,\"Created\":\"%T\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%N,\"Message\":\"%M\"%F}"
	lo := NewLayout(jsonFormat, args...)
	lo.SetOptions("timeEncoder", "rfc3339nano", "fieldsEncoder", "json", "valuesEncoder", "json")
	return lo
}

// NewJSONValueLayout creates a new layout encoding log Recorder as JSON format with Values []interface{}.
func NewJSONValueLayout(args ...interface{}) *PatternLayout {
	jsonFormat := "{\"Level\":%l,\"Created\":\"%T\",\"Prefix\":\"%P\",\"Source\":\"%S\",\"Line\":%N,\"Message\":\"%M\"%V}"
	lo := NewLayout(jsonFormat, args...)
	lo.SetOptions("timeEncoder", "rfc3339nano", "fieldsEncoder", "json", "valuesEncoder", "json")
	return lo
}
