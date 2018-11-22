// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"strings"
	"strconv"
)

type NameValue struct {
	Name  string `xml:"name,attr" json:"name"`
	Value string `xml:",chardata" json:"value"`
}

type FilterConfig struct {
	Enabled string `xml:"enabled,attr" json:"enabled"`
	Tag     string `xml:"tag" json:"tag"`
	Type    string `xml:"type" json:"type"`
	Level   string `xml:"level" json:"level"`
	Pattern string `xml:"format" json:"format"`
	Properties []NameValue `xml:"property" json:"properties"`
}

type LoggerConfig struct {
	Filters []FilterConfig `xml:"filter" json:"filters"`
}

func loadAppender(typ string, props []NameValue) (Appender, []string) {
	var errs []string

	newFunc := GetAppenderNewFunc(typ)
	if newFunc == nil {
		errs = append(errs, "Unknown appender type. " + typ)
		return nil, errs
	}

	appender := newFunc()
	for _, prop := range props {
		v := strings.Trim(prop.Value, " \r\n")
		err := appender.SetOption(prop.Name, v)
		if err != nil {
			errs = append(errs, err.Error() + ". " + prop.Name + ": " + v)
		}
	}
	return appender, errs
}

// Load configuration; see examples/example.xml for documentation
func (log *Logger) LoadConfiguration(lc *LoggerConfig) {
	if lc == nil {
		LogLogWarn("Logger configuration is NIL")
		return
	}
	if len(lc.Filters) <= 0 {
		LogLogTrace("Filters configuration is NIL")
		return
	}

	filters := make(Filters)
	for _, fc := range lc.Filters {
		if fc.Tag == "" && fc.Type == "" {
			LogLogWarn("Missing tag and type")
			continue
		}

		if fc.Tag == "" { 
			fc.Tag = fc.Type
		}
		if fc.Type == "" {
			fc.Type = strings.ToLower(fc.Tag)
		}

		enabled, err := strconv.ParseBool(fc.Enabled)
		if err != nil {
			LogLogTrace("Disable \"%s\" for %s", fc.Tag, err)
			continue
		}
		if !enabled {
			LogLogTrace("Disable \"%s\"", fc.Tag)
			continue
		} 

		if fc.Level == "" {
			fc.Level = levelStrings[INFO]
		}
		lvl := GetLevel(fc.Level)

		if fc.Type == "stdout" {
			if lvl >= SILENT {
				LogLogTrace("Disable \"%s\" for level \"%s\"", fc.Tag, fc.Level)
				log.SetOutput(nil)
			} else {
				log.SetLevel(lvl)
				if fc.Pattern != "" {
					log.SetPattern(fc.Pattern)
				}
			}
			continue
		}

		if lvl >= SILENT {
			LogLogTrace("Disable \"%s\" for level \"%s\"", fc.Tag, fc.Level)
			continue
		}

		if fc.Type == "loglog" {
			loglog := GetLogLog().SetLevel(lvl)
			if fc.Pattern != "" {
				loglog.SetPattern(fc.Pattern)
			}
			continue
		}

		appender, errs := loadAppender(fc.Type, fc.Properties)
		if len(errs) > 0 {
			for _, err := range errs {
				LogLogWarn(err)
			}
			LogLogTrace("Failed loading appender \"%s\"", fc.Tag)
			continue
		}

		LogLogTrace("Succeeded loading appender \"%s\"", fc.Tag)
		filters.Add(fc.Tag, lvl, appender)
	}

	log.SetFilters(filters)
}