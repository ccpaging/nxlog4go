// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package configtest

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"encoding/xml"
	l4g "github.com/ccpaging/nxlog4go"
	_ "github.com/ccpaging/nxlog4go/color"
	_ "github.com/ccpaging/nxlog4go/file"
	_ "github.com/ccpaging/nxlog4go/socket"
)

func TestXMLConfig(t *testing.T) {
	const (
		configfile = "example.xml"
	)

	fd, err := os.Create(configfile)
	if err != nil {
		t.Fatalf("Could not open %s for writing: %s", configfile, err)
	}

	fmt.Fprint(fd, 
`<logging>
  <filter enabled="true">
    <type>stdout</type>
    <!-- level is (:?FINEST|FINE|DEBUG|TRACE|INFO|WARNING|ERROR) -->
    <level>DEBUG</level>
    <!--
      %T - Time (15:04:05 MST)
      %`+`t - Time (15:04)
      %D - Date (2006/01/02)
      %`+`d - Date (01/02/06)
      %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
      %S - Source
      %M - Message
      It ignores unknown format strings (and removes them)
      Recommended: \"[%D %T] [%L] (%S) %M\"
    -->
    <property name="pattern">[%D %T] [%L] (%S) %M</property>
  </filter>
  <filter enabled="true">
    <type>loglog</type>
    <level>DEBUG</level>
    <property name="pattern">[%D %T] [%L] (%S) %M</property>
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
    <property name="filename">test.log</property>
    <property name="pattern">[%D %T] [%L] (%S) %M</property>
    <property name="maxbackup">7</property> <!-- 0, disables log rotation, otherwise append -->
    <property name="maxsize">10M</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->
    <property name="cycle">5m</property> <!-- The cycle time with with fraction and a unit suffix -->
  </filter>
  <filter enabled="true">
    <tag>xmllog</tag>
    <type>xml</type>
    <level>TRACE</level>
    <property name="filename">trace.xml</property>
    <property name="maxbackup">7</property> <!-- 0, disables log rotation, otherwise append -->
    <property name="maxsize">100M</property> <!-- \\d+[KMG]? Suffixes are in terms of 2**10 -->
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
</logging>`)

	fd.Close()

	// Open the configuration file
	fd, err = os.Open(configfile)
	if err != nil {
		t.Fatalf("XMLConfig: Could not open %q for reading: %s\n", configfile, err)
	}

	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		t.Fatalf("XMLConfig: Could not read %q: %s\n", configfile, err)
	}

	lc := new(l4g.LoggerConfig)
	if err := xml.Unmarshal(contents, lc); err != nil {
		t.Fatalf("XMLConfig: Could not parse XML configuration in %q: %s\n", configfile, err)
	}
	
	log := l4g.New(l4g.INFO)
	log.LoadConfiguration(lc)
	filters := log.Filters()

	defer os.Remove("trace.xml")
	defer os.Remove("test.log")
	defer func() {
		log.SetFilters(nil)
		if filters != nil {
			filters.Close()
		}
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
	if filters["color"].Appender == nil {
		t.Fatalf("XMLConfig: Expected stdout to be ConsoleLogWriter, found %T", filters["color"].Appender.Write)
	}
	/*
	if _, ok := filters.Get("file").LogWriter.(*FileLogWriter); !ok {
		t.Fatalf("XMLConfig: Expected file to be *FileLogWriter, found %T", log["file"].LogWriter)
	}
	if _, ok := filters.Get("xmllog").LogWriter.(*FileLogWriter); !ok {
		t.Fatalf("XMLConfig: Expected xmllog to be *FileLogWriter, found %T", log["xmllog"].LogWriter)
	}

	// Make sure levels are set
	if lvl := filters.Get("stdout").Level; lvl != DEBUG {
		t.Errorf("XMLConfig: Expected stdout to be set to level %d, found %d", DEBUG, lvl)
	}
	if lvl := filters.Get("file").Level; lvl != FINEST {
		t.Errorf("XMLConfig: Expected file to be set to level %d, found %d", FINEST, lvl)
	}
	if lvl := filters.Get("xmlog").Level; lvl != TRACE {
		t.Errorf("XMLConfig: Expected xmllog to be set to level %d, found %d", TRACE, lvl)
	}

	// Make sure the w is open and points to the right file
	if fname := filters.Get("file").LogWriter.(*FileLogWriter).file.Name(); fname != "test.log" {
		t.Errorf("XMLConfig: Expected file to have opened %s, found %s", "test.log", fname)
	}

	// Make sure the XLW is open and points to the right file
	if fname := filters.Get("xmllog").LogWriter.(*FileLogWriter).file.Name(); fname != "trace.xml" {
		t.Errorf("XMLConfig: Expected xmllog to have opened %s, found %s", "trace.xml", fname)
	}

	// Move XML log file
	os.Rename(configfile, "examples/"+configfile) // Keep this so that an example with the documentation is available
	*/
}

