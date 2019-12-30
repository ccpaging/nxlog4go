package nxlog4go

import (
	"errors"
	"io"
	"os"

	"github.com/ccpaging/nxlog4go/driver"
	"github.com/ccpaging/nxlog4go/patt"
)

type stda struct {
	out io.Writer // destination for output
}

func (a *stda) Open(dsn string, args ...interface{}) (driver.Appender, error) {
	return &stda{os.Stderr}, nil
}

func (a *stda) Set(name string, value interface{}) error { return nil }
func (a *stda) Enabled(*driver.Recorder) bool            { return true }
func (a *stda) Write(b []byte) (int, error) {
	if a.out == nil {
		return 0, errors.New("io writer is nil")
	}
	return a.out.Write(b)
}
func (a *stda) Close() {}

type stdf struct {
	*driver.Filter

	level int // The log level
	flag  int // properties compatible with go std log

	*stda
}

func newStdf(level int) *stdf {
	f := &stdf{Filter: &driver.Filter{}}
	return f.setLevel(level).setFormat("").setWriter(os.Stderr)
}

func (f *stdf) getCaller() bool {
	if f.flag&(Lshortfile|Llongfile) != 0 {
		return true
	}

	return false
}

func (f *stdf) enable(enb bool) *stdf {
	if !enb {
		f.Enabler = driver.DenyAll()
	} else {
		f.setLevel(f.level)
	}
	return f
}

func (f *stdf) setLevel(level int) *stdf {
	f.level = level
	f.Enabler = driver.AtAbove(level)
	return f
}

func (f *stdf) setWriter(out io.Writer) *stdf {
	if f.stda == nil {
		f.stda = &stda{out}
	}
	if len(f.Apps) == 0 {
		f.Apps = []driver.Appender{f.stda}
	}
	f.stda.out = out
	if out == nil {
		f.Enabler = driver.DenyAll()
	} else {
		f.Enabler = driver.AtAbove(f.level)
	}
	return f
}

func (f *stdf) setFlags(flag int) *stdf {
	f.flag = flag

	if f.Layout == nil {
		f.Layout = patt.NewLayout("%D %T %M")
	}

	if flag&LUTC != 0 {
		f.Layout.Set("utc", true)
	} else {
		f.Layout.Set("utc", false)
	}

	format := "%P"
	if flag&Ldate != 0 {
		format += "%D "
	}
	if flag&(Ltime|Lmicroseconds) != 0 {
		format += "%T "
		if flag&Lmicroseconds != 0 {
			f.Layout.Set("timeEncoder", "hms.us")
		} else {
			f.Layout.Set("timeEncoder", "hms")
		}
	}
	if flag&(Lshortfile|Llongfile) != 0 {
		format += "%S:%N: "
		if flag&Lshortfile != 0 {
			f.Layout.Set("callerEncoder", "nopath")
		} else {
			f.Layout.Set("callerEncoder", "fullpath")
		}
	}
	format += "%M"
	f.Layout.Set("format", format)
	return f
}

func (f *stdf) setFormat(format string) *stdf {
	if f.Layout == nil {
		f.Layout = patt.NewLayout(format)
	} else {
		f.Layout.Set("format", format)
	}
	return f
}
