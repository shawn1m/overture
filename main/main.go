// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core"
)

var version string

func main() {

	var (
		configPath      string
		isLogVerbose    bool
		processorNumber int
	)

	flag.StringVar(&configPath, "c", "./config.json", "config file path")
	flag.BoolVar(&isLogVerbose, "v", false, "verbose mode")
	flag.IntVar(&processorNumber, "p", runtime.NumCPU(), "number of processor to use")
	flag.Parse()

	if isLogVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Info("Overture " + version)
	log.Info("If you need any help or want to check update, please visit the project repository: https://github.com/shawn1m/overture")

	runtime.GOMAXPROCS(processorNumber)

	core.Init(configPath)
}
