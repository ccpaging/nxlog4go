// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package file

import (
	"testing"

	"github.com/ccpaging/nxlog4go/cast"
)

/* default log file is "file.log" in linux
func TestFileName(t *testing.T) {
	a, _ := NewAppender("")
	if a.out == nil {
		t.Errorf("out is nil")
	}

	want := "file.test.log"
	if got := a.out.file.Name; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
*/

func TestRotateOption(t *testing.T) {
	var v interface{} = "true"

	if rotate, err := cast.ToInt(v); err == nil {
		t.Errorf("rotate %q = %d", v, rotate)
	}

	if _, err := cast.ToBool(v); err != nil {
		t.Errorf("convert rotate %q to bool. %v", v, err)
	}
}
