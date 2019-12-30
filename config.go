// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"strings"

	"github.com/ccpaging/nxlog4go/cast"
	"github.com/ccpaging/nxlog4go/driver"
)

// NameValue stores every single option's name and value.
type NameValue struct {
	Name  string `xml:"name,attr" json:"name"`
	Value string `xml:",chardata" json:"value"`
}

// FilterConfig offers a declarative way to construct a logger's default writer,
// internal log and 3rd appenders
type FilterConfig struct {
	Enabled    string      `xml:"enabled,attr" json:"enabled"`
	Tag        string      `xml:"tag" json:"tag"`
	Type       string      `xml:"type" json:"type"`
	Dsn        string      `xml:"dsn" json:"dsn"`
	Level      string      `xml:"level" json:"level"`
	Properties []NameValue `xml:"property" json:"properties"`
}

// LoggerConfig offers a declarative way to construct a logger.
// See examples/config.xml and examples/config.json for documentation
type LoggerConfig struct {
	Version string          `xml:"version,attr" json:"version"`
	Filters []*FilterConfig `xml:"filter" json:"filters"`
}

func setLogger(l *Logger, enb bool, fc *FilterConfig) (errs []error) {
	l.Set("level", Level(INFO).Int(fc.Level))
	for _, prop := range fc.Properties {
		v := strings.Trim(prop.Value, " \r\n")
		if err := l.Set(prop.Name, v); err != nil {
			errs = append(errs, fmt.Errorf("Warn: %s. %s: %s", err.Error(), prop.Name, v))
		}
	}
	l.Enable(enb)
	return
}

func loadFilter(fc *FilterConfig) (filter *driver.Filter, errs []error) {
	app, err := driver.Open(fc.Type, fc.Dsn)
	if app == nil {
		return nil, append(errs, err)
	}

	for _, prop := range fc.Properties {
		v := strings.Trim(prop.Value, " \r\n")
		if err := app.Set(prop.Name, v); err != nil {
			errs = append(errs, fmt.Errorf("Warn: Set [%s] as [%s=%s]. %s", fc.Type, prop.Name, v, err.Error()))
		}
	}

	filter = &driver.Filter{
		Name:    fc.Tag,
		Enabler: driver.AtAbove(Level(INFO).Int(fc.Level)),
		Layout:  nil,
		Apps:    []driver.Appender{app},
	}

	errs = append(errs, fmt.Errorf("Trace: Succeeded loading tag [%s], type [%s], dsn [%s]", fc.Tag, fc.Type, fc.Dsn))
	return
}

// LoadConfiguration sets options of logger, and creates/loads/sets appenders.
func (l *Logger) LoadConfiguration(lc *LoggerConfig) (errs []error) {
	if lc == nil {
		return append(errs, fmt.Errorf("Warn: Logger configuration is NIL"))
	}

	var filters []*driver.Filter
	for i, fc := range lc.Filters {
		if fc.Type == "" {
			errs = append(errs, fmt.Errorf("Warn: The type of Filter [%d] is not defined", i))
			continue
		}
		if fc.Tag == "" {
			fc.Tag = fc.Type

		}

		enabled, err := cast.ToBool(fc.Enabled)
		if err != nil {
			errs = append(errs, fmt.Errorf("Trace: Disable filter [%s]. Error: %v", fc.Tag, err))
		} else if !enabled {
			errs = append(errs, fmt.Errorf("Trace: Disable filter [%s]", fc.Tag))
		}

		var e []error
		if fc.Type == "loglog" {
			e = setLogger(GetLogLog(), enabled, fc)
		} else if fc.Type == stdfName {
			e = setLogger(l, enabled, fc)
		} else if enabled {
			var filter *driver.Filter
			filter, e = loadFilter(fc)
			if filter != nil {
				filters = append(filters, filter)
			}
		} else {
			// disabled
		}
		errs = append(errs, e...)
	}

	l.Attach(filters...)
	return
}
