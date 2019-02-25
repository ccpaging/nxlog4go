// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"fmt"
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

func setLogger(l *Logger, level Level, props []NameValue) (errs []error) {
	if level >= SILENT {
		return append(errs, fmt.Errorf("Trace: Disable stdout for level [%d]", level))
	}

	l.Set("level", level)
	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		if err := l.SetOption(prop.Name, v); err != nil {
			errs = append(errs, fmt.Errorf("Warn: %s. %s: %s", err.Error(), prop.Name, v))
		}
	}
	return
}

func loadAppender(level Level, typ string, props []NameValue) (app Appender, errs []error) {
	if level >= SILENT {
		return nil, append(errs, fmt.Errorf("Trace: Disable appender type [%s] for level [%d]", typ, level))
	}

	newFunc := getAppenderNewFunc(typ)
	if newFunc == nil {
		return nil, append(errs, fmt.Errorf("Warn: Unknown appender type [%s]", typ))
	}

	app = newFunc()
	if app == nil {
		return nil, append(errs, fmt.Errorf("Warn: Can not create appender type [%s]", typ))
	}

	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		if err := app.SetOption(prop.Name, v); err != nil {
			errs = append(errs, fmt.Errorf("Warn: %s. %s: %s", err.Error(), prop.Name, v))
		}
	}
	return
}

// LoadConfiguration sets options of logger, and creates/loads/sets appenders.
func (l *Logger) LoadConfiguration(lc *LoggerConfig) (errs []error) {
	if lc == nil {
		return append(errs, fmt.Errorf("Warn: Logger configuration is NIL"))
	}
	filters := make(Filters)
	for i, fc := range lc.Filters {
		if fc.Type == "" {
			errs = append(errs, fmt.Errorf("Warn: The type of Filter [%d] is not defined", i))
			continue
		}
		if fc.Tag == "" {
			errs = append(errs, fmt.Errorf("Warn: The tag of Filter [%d] is not defined", i))
			continue
		}

		if enabled, err := ToBool(fc.Enabled); !enabled {
			errs = append(errs, fmt.Errorf("Trace: Disable filter [%s]. Error: %v", fc.Tag, err))
			continue
		}

		level := GetLevel(fc.Level)

		switch fc.Type {
		case "loglog":
			retErrors := setLogger(GetLogLog(), level, fc.Properties)
			errs = append(errs, retErrors...)
		case "stdout":
			retErrors := setLogger(l, level, fc.Properties)
			errs = append(errs, retErrors...)
		default:
			app, retErrors := loadAppender(level, fc.Type, fc.Properties)
			errs = append(errs, retErrors...)
			if app != nil {
				errs = append(errs, fmt.Errorf("Trace: Succeeded loading appender [%s]", fc.Tag))
				filters.Add(fc.Tag, level, app)
			}
		}
	}

	l.SetFilters(filters)
	return
}
