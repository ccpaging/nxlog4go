// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package nxlog4go

import (
	"bytes"
	"strings"
)

var (
	FORMAT_DEFAULT = "[%D %T %z] [%L] (%s) %M"
	FORMAT_SHORT   = "[%t %d] [%L] %M"
	FORMAT_ABBREV  = "[%L] %M"
	FORMAT_UTC	   = false
)

type formatCacheType struct {
	LastUpdateSeconds   int64
	longTime, shortTime []byte
	longZone, shortZone []byte
	longDate, shortDate []byte
}

var formatCache = &formatCacheType{}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// Known format codes:
// %N - Time (15:04:05.000000)
// %T - Time (15:04:05)
// %t - Time (15:04)
// %Z - Zone (-0700)
// %z - Zone (MST)
// %D - Date (2006/01/02)
// %d - Date (01/02/06)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %P - Prefix
// %S - Source
// %s - Short Source
// %x - Extra Short Source: just file without .go suffix
// %M - Message
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
func FormatLogRecord(formatSlice [][]byte, rec *LogRecord) []byte {
	if rec == nil {
		return []byte("<nil>")
	}
	if len(formatSlice) == 0 {
		return []byte{}
	}

	t := rec.Created
	if FORMAT_UTC {
		t = t.UTC()
	}

	out := bytes.NewBuffer(make([]byte, 0, 64))
	secs := t.UnixNano() / 1e9

	cache := *formatCache
	if cache.LastUpdateSeconds != secs {
		year, month, day := t.Date()
		hour, minute, second := t.Clock()
		updated := &formatCacheType{
			LastUpdateSeconds: secs,
			shortTime:         make([]byte, 0, 16),
			longTime:          make([]byte, 0, 16),
			shortZone:         []byte(t.Format("MST")),
			longZone:          []byte(t.Format("-0700")),
			shortDate:         make([]byte, 0, 16),
			longDate:   	   make([]byte, 0, 16),          
		}
		// fmt.Sprintf("%02d:%02d", hour, minute)
		buf := &updated.shortTime 
		itoa(buf, hour, 2); *buf = append(*buf, ':')
 		itoa(buf, minute, 2)
		// fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
		buf = &updated.longTime 
		itoa(buf, hour, 2); *buf = append(*buf, ':')
 		itoa(buf, minute, 2); *buf = append(*buf, ':')
		itoa(buf, second, 2)
		// fmt.Sprintf("%02d/%02d/%02d", day, month, year%100)
		buf = &updated.shortDate
		itoa(buf, day, 2); *buf = append(*buf, '/')
 		itoa(buf, int(month), 2); *buf = append(*buf, '/')
		itoa(buf, year%100, 2)
		// fmt.Sprintf("%04d/%02d/%02d", year, month, day),
		buf = &updated.longDate 
		itoa(buf, year, 4); *buf = append(*buf, '/')
		itoa(buf, int(month), 2); *buf = append(*buf, '/')
		itoa(buf, day, 2)
		
		cache = *updated
		formatCache = updated
	}

	// Split the string into pieces by % signs
	//pieces := bytes.Split([]byte(format), []byte{'%'})
	b := make([]byte, 0, 16)
	// Iterate over the pieces, replacing known formats
	for i, piece := range formatSlice {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'N':
				out.Write(cache.longTime)
				b = b[:0]
				b = append(b, '.')
				itoa(&b, t.Nanosecond()/1e3, 6)
				out.Write(b)
			case 'T':
				out.Write(cache.longTime)
			case 't':
				out.Write(cache.shortTime)
			case 'Z':
				out.Write(cache.longZone)
			case 'z':
				out.Write(cache.shortZone)
			case 'D':
				out.Write(cache.longDate)
			case 'd':
				out.Write(cache.shortDate)
			case 'L':
				out.WriteString(levelStrings[rec.Level])
			case 'P':
                out.WriteString(rec.Prefix)
			case 'S':
				out.WriteString(rec.Source)
				if rec.Line > 0 {
					b = b[:0]
					b = append(b, ':')
					itoa(&b, rec.Line, -1)
					out.Write(b)
				}
			case 's':
				out.WriteString(rec.Source[strings.LastIndexByte(rec.Source, '/')+1:])
				if rec.Line > 0 {
					b = b[:0]
					b = append(b, ':')
					itoa(&b, rec.Line, -1)
					out.Write(b)
				}
			case 'M':
				out.WriteString(rec.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.Bytes()
}

