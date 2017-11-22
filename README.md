# Log [![Go Test](https://travis-ci.org/micanzhang/log.svg?branch=master)](https://travis-ci.org/micanzhang/log.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/micanzhang/log)](https://goreportcard.com/report/github.com/micanzhang/log)

## FluentdFormatter

Useful addition to logrus, allowing it to format log entries that can be parsed by Kubernetes
and Google Container Engine.

Example:

```go
package main

import (
	"os"
	"fmt"
	"flag"

	log "github.com/sirupsen/logrus"
	joonix "github.com/joonix/log"
)

func main() {
	lvl := flag.String("level", log.DebugLevel.String(), "log level")
	flag.Parse()

	level, err := log.ParseLevel(*lvl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.SetLevel(level)
	log.SetFormatter(&joonix.FluentdFormatter{})

	log.Debug("hello world!")		
}
```
