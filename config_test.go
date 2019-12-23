// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	l4g "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/console"
	_ "github.com/ccpaging/nxlog4go/file"
	_ "github.com/ccpaging/nxlog4go/socket"
	"io/ioutil"
	"os"
	"testing"
)

var xmlBuf = `<logging>
  <filter enabled="true">
    <tag>stdout</tag>
    <type>stdout</type>
    <level>DEBUG</level>
    <property name="format">[%D %T] [%L] (%S) %M</property>
    <property name="color">true</property>
  </filter>
  <filter enabled="true">
    <tag>loglog</tag>
    <type>loglog</type>
    <level>DEBUG</level>
    <property name="format">[%D %T] [%L] (%S) %M</property>
  </filter>
  <filter enabled="true">
    <tag>console</tag>
    <type>console</type>
    <!-- level is (:?FINEST|FINE|DEBUG|TRACE|INFO|WARN|ERROR) -->
    <level>DEBUG</level>
    <property name="color">true</property>
  </filter>
  <filter enabled="true">
    <tag>file</tag>
    <type>file</type>
    <level>FINEST</level>
    <property name="filename">_test.log</property>
    <!--
      %T - Time (15:04:05 MST)
      %t - Time (15:04)
      %D - Date (2006/01/02)
      %d - Date (01/02/06)
      %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
      %S - Source
      %M - Message
      It ignores unknown format strings (and removes them)
      Recommended: "[%D %T] [%L] (%S) %M"
    -->
    <property name="format">[%D %T] [%L] (%S) %M</property>
    <property name="rotate">false</property> <!-- true enables log rotation, otherwise append -->
    <property name="maxsize">0M</property> <!-- \d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="maxlines">0K</property> <!-- \d+[KMG]? Suffixes are in terms of thousands -->
    <property name="daily">true</property> <!-- Automatically rotates when a log message is written after midnight -->
  </filter>
  <filter enabled="true">
    <tag>xmlog</tag>
    <type>xml</type>
    <level>TRACE</level>
    <property name="filename">_trace.xml</property>
    <property name="rotate">true</property> <!-- true enables log rotation, otherwise append -->
    <property name="maxsize">100M</property> <!-- \d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="maxrecords">6K</property> <!-- \d+[KMG]? Suffixes are in terms of thousands -->
    <property name="daily">false</property> <!-- Automatically rotates when a log message is written after midnight -->
  </filter>
  <filter enabled="false"><!-- enabled=false means this logger won't actually be created -->
    <tag>donotopen</tag>
    <type>socket</type>
    <level>FINEST</level>
    <property name="endpoint">192.168.1.255:12124</property> <!-- recommend UDP broadcast -->
    <property name="protocol">udp</property> <!-- tcp or udp -->
  </filter>
</logging>`

var xmlFile = "examples/example.xml"
var jsonFile = "examples/example.json"

func TestXMLConfig(t *testing.T) {
	fd, err := os.Create(xmlFile)
	if err != nil {
		t.Fatalf("Could not open %s for writing: %s", xmlFile, err)
	}

	fmt.Fprint(fd, xmlBuf)

	fd.Close()

	// Open the configuration file
	fd, err = os.Open(xmlFile)
	if err != nil {
		t.Fatalf("XMLConfig: Could not open %q for reading: %s\n", xmlFile, err)
	}

	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatalf("XMLConfig: Could not read %q: %s\n", xmlFile, err)
	}

	lc := new(l4g.LoggerConfig)
	if err := xml.Unmarshal(contents, lc); err != nil {
		t.Fatalf("XMLConfig: Could not parse XML configuration in %q: %s\n", xmlFile, err)
	}

	log := l4g.NewLogger(l4g.DEBUG)
	log.LoadConfiguration(lc)
	filters := log.Filters()

	defer func() {
		log.Close()
		os.Remove("_trace.xml")
		os.Remove("_test.log")
	}()

	// Make sure we got all loggers
	if filters == nil {
		t.Fatalf("XMLConfig: Expected 3 filters, found %d", len(filters))
	}

	if len(filters) != 3 {
		t.Fatalf("XMLConfig: Expected 3 filters, found %d", len(filters))
	}

	// Make sure they're the right type
	for i, filter := range filters {
		if fmt.Sprintf("%T", filter.Dispatch) != "func(*driver.Recorder)" {
			t.Fatalf("XMLConfig: Expected [%d] filter Dispatch(*nxlog4go.Recorder), found %T", i, filter.Dispatch)
		}
	}

	// Make sure levels are set
	/*
		if level := filters[0].level; level != l4g.DEBUG {
			t.Errorf("XMLConfig: Expected stdout to be set to level %d, found %d", l4g.DEBUG, level)
		}
		if level := filters[1].level; level != l4g.FINEST {
			t.Errorf("XMLConfig: Expected file to be set to level %d, found %d", l4g.FINEST, level)
		}
		if level := filters[2].level; level != l4g.TRACE {
			t.Errorf("XMLConfig: Expected xmlog to be set to level %d, found %d", l4g.TRACE, level)
		}
	*/

	// Keep xmlFile so that an example with the documentation is available

	// Create xmlFile so that an example with the documentation is available

	jsonBuf, _ := json.MarshalIndent(lc, "", "    ")

	fd, err = os.Create(jsonFile)
	if err != nil {
		t.Fatalf("Could not open %s for writing: %s", jsonFile, err)
	}
	fmt.Fprint(fd, string(jsonBuf))
	fd.Close()
}
