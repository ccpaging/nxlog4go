// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"strings"
)

type dict struct {
	data  map[string]string
	index []string
}

func (d *dict) set(name, value string) {
	if d.data == nil {
		d.data = make(map[string]string)
	}

	if _, ok := d.data[name]; ok {
		d.data[name] = value
		return
	}

	d.data[name] = value
	d.index = append(d.index, name)
}

func (d *dict) unset(key string) {
	for i, name := range d.index {
		if name == key {
			d.index = append(d.index[:i], d.index[i+1:]...)
		}
	}
}

func (fc *FilterConfig) toDict() *dict {
	d := new(dict)
	for _, prop := range fc.Properties {
		d.set(prop.Name, strings.Trim(prop.Value, " \r\n"))
	}
	return d
}

func (fc *FilterConfig) fromDict(d *dict) *FilterConfig {
	fc.Properties = nil
	for _, name := range d.index {
		if _, ok := d.data[name]; ok {
			fc.Properties = append(fc.Properties, NameValue{name, d.data[name]})
		}
	}
	return fc
}

func (lc *LoggerConfig) Upgrade() *LoggerConfig {
	if lc.Version != "2.0.0" {
		lc = lc.upgradeToVersion2()
	}
	return lc
}

func (lc *LoggerConfig) upgradeToVersion2() *LoggerConfig {
	lc.Version = "2.0.0"
	for _, fc := range lc.Filters {
		d := fc.toDict()

		if fc.Type == "file" || fc.Type == "xml" {
			if fc.Dsn == "" {
				if filename, ok := d.data["filename"]; ok {
					fc.Dsn = filename
					d.unset("filename")
				}
			}
		} else if fc.Type == "socket" {
			if fc.Dsn == "" {
				protocol, _ := d.data["protocol"]
				if protocol == "" {
					protocol = "udp"
				}
				endpoint, _ := d.data["endpoint"]
				fc.Dsn = protocol + "://" + endpoint

				d.unset("protocol")
				d.unset("endpoint")
			}
		}

		if format, ok := d.data["format"]; ok {
			if strings.Contains(format, "%d") {
				format = strings.Replace(format, "%d", "%D", -1)
				d.set("format", format)
				d.set("dateEncoder", "dmy")
			}
			if strings.Contains(format, "%t") {
				format = strings.Replace(format, "%t", "%T", -1)
				d.set("format", format)
				d.set("timeEncoder", "hhmm")
			}
		}

		if enableRotate, ok := d.data["rotate"]; ok {
			if strings.ToUpper(enableRotate) == "TRUE" {
				d.set("rotate", "1")
			} else {
				d.set("rotate", "-1")
			}
		}

		if daily, ok := d.data["daily"]; ok {
			if strings.ToUpper(daily) == "TRUE" {
				d.set("cycle", "86400s")
				d.set("clock", "0s")
				d.set("maxsize", "5k")
			} else {
				d.set("cycle", "5s")
				d.set("clock", "0s")
			}
			d.unset("daily")
		}

		fc.fromDict(d)
	}
	return lc
}
