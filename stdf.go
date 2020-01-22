package nxlog4go

import (
	"bytes"
	"io"
	"os"

	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

type stdFilter struct {
	level int // The log level
	flag  int // properties compatible with go std log
	enb   bool
	lo    driver.Layout
	out   io.Writer // destination for output
}

func newStdFilter(level int) *stdFilter {
	return &stdFilter{
		level: level,
		enb:   true,
		lo:    patt.NewLayout(""),
		out:   os.Stderr,
	}
}

func (f *stdFilter) setFlags(flag int) *stdFilter {
	f.flag = flag

	if flag&LUTC != 0 {
		f.lo.Set("utc", true)
	} else {
		f.lo.Set("utc", false)
	}

	format := "%P"
	if flag&Ldate != 0 {
		format += "%D "
	}
	if flag&(Ltime|Lmicroseconds) != 0 {
		format += "%T "
		if flag&Lmicroseconds != 0 {
			f.lo.Set("timeEncoder", "hms.us")
		} else {
			f.lo.Set("timeEncoder", "hms")
		}
	}
	if flag&(Lshortfile|Llongfile) != 0 {
		format += "%S:%N: "
		if flag&Lshortfile != 0 {
			f.lo.Set("callerEncoder", "nopath")
		} else {
			f.lo.Set("callerEncoder", "fullpath")
		}
	}
	format += "%M"
	f.lo.Set("format", format)
	return f
}

func (f *stdFilter) enabled(level int) bool {
	if f.enb && f.out != nil && level >= f.level {
		return true
	}
	return false
}

func (f *stdFilter) dispatch(r *driver.Recorder) {
	buf := new(bytes.Buffer)
	f.lo.Encode(buf, r)
	f.out.Write(buf.Bytes())
}
