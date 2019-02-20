# nxlog4go - Next logging package based on log4go 
[![License](https://img.shields.io/pypi/l/Django.svg)](https://github.com/ccpaging/nxlog4go/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/ccpaging/nxlog4go?status.svg)](https://godoc.org/github.com/ccpaging/nxlog4go) [![Build Status](https://travis-ci.org/ccpaging/nxlog4go.svg?branch=master)](https://travis-ci.org/ccpaging/nxlog4go) [![Maintainability](https://codeclimate.com/github/ccpaging/nxlog4go/badges/gpa.svg)](https://codeclimate.com/github/ccpaging/nxlog4go/maintainability)

## Installation

- Run `go get github.com/ccpaging/nxlog4go`

OR

- Run `go install github.com/ccpaging/nxlog4go`

## Features:

* **Structured**
* **Fast** - see benchmark
* **Extendable 3rd appender**
* **Configurable**
* **Robust**
* **No external dependencies** - plain ol' Go stdlib
* **Completely API compatible with the stdlib log** - support ansi color term

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
}
```

## Acknowledgements:

* <https://github.com/alecthomas/log4go/>
* <https://github.com/ngmoco/timber>
* <https://github.com/siddontang/go/tree/master/log>
* <https://github.com/Sirupsen/logrus>
* <https://github.com/YoungPioneers/blog4go>
* <https://github.com/cihub/seelog>
* <https://github.com/golang/glog>
* <https://github.com/go-chi/chi/blob/master/middleware/terminal.go>

## Stability

This package is a work still in progress. 

In any event, I encourage you to send me feedback.

But I really wouldn't recommend using nxlog4go in factory environment just yet.
