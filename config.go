// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"strings"
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

func loadLogLog(l Level, pattern string) {
	if l < SILENT {
		loglog := GetLogLog().SetLevel(l)
		if pattern != "" {
			loglog.SetPattern(pattern)
		}
	}
}

func loadStdout(log *Logger, l Level, pattern string) {
	if l < SILENT {
		log.SetLevel(l)
		if pattern != "" {
			log.SetPattern(pattern)
		}
	} else {
		log.SetOutput(nil)
	}
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
		} else if fc.Tag == "" { 
			fc.Tag = fc.Type
		} else if fc.Type == "" {
			fc.Type = strings.ToLower(fc.Tag)
		}

		if enabled, err := ToBool(fc.Enabled); err != nil {
			LogLogTrace("Disable \"%s\" for %s", fc.Tag, err)
			continue
		} else if !enabled {
			LogLogTrace("Disable \"%s\"", fc.Tag)
			continue
		} 

		level := GetLevel(fc.Level)
		if level >= SILENT {
			LogLogTrace("Disable \"%s\" for level \"%s\"", fc.Tag, fc.Level)
		}

		switch fc.Type {
		case "loglog":
			loadLogLog(level, fc.Pattern)
		case "stdout":
			loadStdout(log, level, fc.Pattern)
		default:
			if level >= SILENT {
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
			filters.Add(fc.Tag, level, appender)
		}
	}

	log.SetFilters(filters)
}