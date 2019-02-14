// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	l4g "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/color"
	"github.com/ccpaging/nxlog4go/file"
	_ "github.com/ccpaging/nxlog4go/socket"
	"io/ioutil"
	"os"
	"testing"
)

var xmlBuf = `<logging>
  <filter enabled="true">
    <tag>stdout</tag>
    <type>stdout</type>
    <!-- level is (:?FINEST|FINE|DEBUG|TRACE|INFO|WARNING|ERROR) -->
    <level>DEBUG</level>
    <!--
      %T - Time (15:04:05 MST)
      %` + `t - Time (15:04)
      %D - Date (2006/01/02)
      %` + `d - Date (01/02/06)
      %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
      %S - Source
      %M - Message
      It ignores unknown format strings (and removes them)
      Recommended: \"[%D %T] [%L] (%S) %M%R\"
    -->
    <property name="pattern">[%D %T] [%L] (%s) %M</property>
  </filter>
  <filter enabled="true">
    <tag>color</tag>
    <type>loglog</type>
    <level>DEBUG</level>
    <property name="pattern">[%D %T] [%L] (%s) %M%R</property>
  </filter>
  <filter enabled="true">
    <tag>color</tag>
    <type>color</type>
    <level>DEBUG</level>
  </filter>
  <filter enabled="true">
    <tag>file</tag>
    <type>file</type>
    <level>FINEST</level>
    <property name="filename">_test.log</property>
    <property name="pattern">[%D %T] [%L] (%s) %M%R</property>
    <property name="maxbackup">7</property> <!-- 0, disables log rotation, otherwise append -->
    <property name="maxsize">10M</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="maxlines">0</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="cycle">5m</property> <!-- The cycle time with with fraction and a unit suffix -->
  </filter>
  <filter enabled="true">
    <tag>xmllog</tag>
    <type>xml</type>
    <level>TRACE</level>
    <property name="filename">_trace.xml</property>
    <property name="maxbackup">7</property> <!-- 0, disables log rotation, otherwise append -->
    <property name="maxsize">0</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="maxlines">100</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="cycle">24h</property> <!-- The cycle time with with fraction and a unit suffix -->
    <property name="clock">0</property> <!-- The cycle time with with fraction and a unit suffix -->
  </filter>
  <filter enabled="false"><!-- enabled=false means this logger won\'t actually be created -->
    <tag>socket</tag>
    <type>socket</type>
    <level>FINEST</level>
    <property name="endpoint">192.168.1.255:12124</property> <!-- recommend UDP broadcast -->
    <property name="protocol">udp</property> <!-- tcp or udp -->
  </filter>
</logging>`

var xmlFile = "examples/config.xml"
var jsonFile = "examples/config.json"

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

	log := new(l4g.Logger)
	log.LoadConfiguration(lc)
	filters := log.Filters()

	defer func() {
		log.SetFilters(nil)
		if filters != nil {
			filters.Close()
		}
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

	// Make sure they're the right keys
	if _, ok := filters["color"]; !ok {
		t.Errorf("XMLConfig: Expected color appender")
	}
	if _, ok := filters["file"]; !ok {
		t.Fatalf("XMLConfig: Expected file appender")
	}
	if _, ok := filters["xmllog"]; !ok {
		t.Fatalf("XMLConfig: Expected xmllog appender")
	}

	// Make sure they're the right type
	if fmt.Sprintf("%T", filters["color"].Write) != "func(*nxlog4go.LogRecord)" {
		t.Fatalf("XMLConfig: Expected color log write, found %T", filters["color"].Write)
	}
	if fmt.Sprintf("%T", filters["file"].Write) != "func(*nxlog4go.LogRecord)" {
		t.Fatalf("XMLConfig: Expected file log write, found %T", filters["file"].Write)
	}
	if fmt.Sprintf("%T", filters["xmllog"].Write) != "func(*nxlog4go.LogRecord)" {
		t.Fatalf("XMLConfig: Expected xmllog log write, found %T", filters["xmllog"].Write)
	}
	// Make sure levels are set
	if level := filters["color"].Level; level != l4g.DEBUG {
		t.Errorf("XMLConfig: Expected stdout to be set to level %d, found %d", l4g.DEBUG, level)
	}
	if level := filters["file"].Level; level != l4g.FINEST {
		t.Errorf("XMLConfig: Expected file to be set to level %d, found %d", l4g.FINEST, level)
	}
	if level := filters["xmllog"].Level; level != l4g.TRACE {
		t.Errorf("XMLConfig: Expected xmllog to be set to level %d, found %d", l4g.TRACE, level)
	}

	// Make sure the w is open and points to the right file
	if fname := filters["file"].Appender.(*filelog.FileAppender).Name(); fname != "_test.log" {
		t.Errorf("XMLConfig: Expected file to have opened %s, found %s", "_test.log", fname)
	}

	// Make sure the XLW is open and points to the right file
	if fname := filters["xmllog"].Appender.(*filelog.FileAppender).Name(); fname != "_trace.xml" {
		t.Errorf("XMLConfig: Expected xmllog to have opened %s, found %s", "_trace.xml", fname)
	}

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
