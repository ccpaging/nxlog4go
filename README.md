# nxlog4go - Next logging package based on log4go 

Installation:

- Run `go get github.com/ccpaging/nxlog4go`

OR

- Run `go install github.com/ccpaging/nxlog4go`

Usage:

- Add the following import:

import log "github.com/ccpaging/nxlog4go"

- Sample, using global log

```
package main

import (
	log "github.com/ccpaging/nxlog4go"
)

func main() {
    log.Debug("This is Debug")
    log.Info("This is Info")

    // Compatibility with `log`
    log.Print("This is Print()")
    log.Println("This is Println()")
    log.Panic("This is Panic()")
}
```

## Acknowledgements:

1. <https://github.com/alecthomas/log4go/>
2. <https://github.com/ngmoco/timber>
3. <https://github.com/siddontang/go/tree/master/log>
4. <https://github.com/sirupsen/logrus>
5. <https://github.com/YoungPioneers/blog4go>
6. <https://github.com/cihub/seelog>
7. <https://github.com/golang/glog>


## Stability

This package is a work still in progress. 

In any event, I encourage you to send me feedback.

But I really wouldn't recommend using nxlog4go just yet.
