// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
	"strings"

	"github.com/ccpaging/nxlog4go/cast"
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

func setLogger(l *Logger, level int, props []NameValue) (errs []error) {
	l.Set("level", level)
	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		if err := l.SetOption(prop.Name, v); err != nil {
			errs = append(errs, fmt.Errorf("Warn: %s. %s: %s", err.Error(), prop.Name, v))
		}
	}
	return
}

func loadFilter(level int, typ string, dsn string, props []NameValue) (filter *Filter, errs []error) {
	app, err := Open(typ, dsn)
	if app == nil {
		return nil, append(errs, err)
	}

	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		if err := app.SetOption(prop.Name, v); err != nil {
			errs = append(errs, fmt.Errorf("Warn: Set [%s] as [%s=%s]. %s", typ, prop.Name, v, err.Error()))
		}
	}

	filter = NewFilter(level, nil, app)
	return
}

// LoadConfiguration sets options of logger, and creates/loads/sets appenders.
func (l *Logger) LoadConfiguration(lc *LoggerConfig) (errs []error) {
	if lc == nil {
		return append(errs, fmt.Errorf("Warn: Logger configuration is NIL"))
	}

	var filters []*Filter
	for i, fc := range lc.Filters {
		if fc.Type == "" {
			errs = append(errs, fmt.Errorf("Warn: The type of Filter [%d] is not defined", i))
			continue
		}
		if fc.Tag == "" {
			fc.Tag = fc.Type

		}

		if enabled, err := cast.ToBool(fc.Enabled); !enabled {
			errs = append(errs, fmt.Errorf("Trace: Disable filter [%s]. Error: %v", fc.Tag, err))
			continue
		}

		level := Level(INFO).Int(fc.Level)

		switch fc.Type {
		case "loglog":
			multiErr := setLogger(GetLogLog(), level, fc.Properties)
			errs = append(errs, multiErr...)
		case "stdout":
			multiErr := setLogger(l, level, fc.Properties)
			errs = append(errs, multiErr...)
		default:
			filter, multiErr := loadFilter(level, fc.Type, fc.Dsn, fc.Properties)
			errs = append(errs, multiErr...)
			if filter != nil {
				errs = append(errs, fmt.Errorf("Trace: Succeeded loading tag [%s], type [%s], dsn [%s]", fc.Tag, fc.Type, fc.Dsn))
				filters = append(filters, filter)
			}
		}
	}

	l.Attach(filters...)
	return
}
