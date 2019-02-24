// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"strings"
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
	Level      string      `xml:"level" json:"level"`
	Properties []NameValue `xml:"property" json:"properties"`
}

// LoggerConfig offers a declarative way to construct a logger.
// See examples/config.xml and examples/config.json for documentation
type LoggerConfig struct {
	Filters []FilterConfig `xml:"filter" json:"filters"`
}

func setLogger(l *Logger, level Level, props []NameValue) {
	if level >= SILENT {
		LogLogTrace("Disable stdout for level \"%d\"", level)
		return
	}

	l.Set("level", level)
	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		if err := l.SetOption(prop.Name, v); err != nil {
			LogLogWarn("%s. %s: %s", err.Error(), prop.Name, v)
		}
	}
}

func loadAppender(level Level, typ string, props []NameValue) Appender {
	if level >= SILENT {
		LogLogTrace("Disable \"%s\" for level \"%d\"", typ, level)
		return nil
	}

	newFunc := getAppenderNewFunc(typ)
	if newFunc == nil {
		LogLogWarn("Unknown appender type \"%s\"", typ)
		return nil
	}

	appender := newFunc()
	if appender == nil {
		return nil
	}

	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		if err := appender.SetOption(prop.Name, v); err != nil {
			LogLogWarn("%s. %s: %s", err.Error(), prop.Name, v)
		}
	}
	return appender
}

// LoadConfiguration sets options of logger, and creates/loads/sets appenders.
func (l *Logger) LoadConfiguration(lc *LoggerConfig) {
	if lc == nil {
		LogLogWarn("Logger configuration is NIL")
		return
	}
	filters := make(Filters)
	for _, fc := range lc.Filters {
		if fc.Type == "" {
			LogLogWarn("Missing type")
			continue
		}
		if fc.Tag == "" {
			LogLogWarn("Missing tag")
			continue
		}

		if enabled, err := ToBool(fc.Enabled); !enabled {
			LogLogTrace("Disable \"%s\" for %v", fc.Tag, err)
			continue
		}

		level := GetLevel(fc.Level)

		switch fc.Type {
		case "loglog":
			setLogger(GetLogLog(), level, fc.Properties)
		case "stdout":
			setLogger(l, level, fc.Properties)
		default:
			appender := loadAppender(level, fc.Type, fc.Properties)
			if appender != nil {
				LogLogTrace("Succeeded loading appender \"%s\"", fc.Tag)
				filters.Add(fc.Tag, level, appender)
			}
		}
	}

	l.SetFilters(filters)
}
