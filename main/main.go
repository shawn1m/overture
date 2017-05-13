// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package main is the entry point of whole program.
package main

import (
	"flag"
	"io"
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core"
)

// For auto version building
//  go build -ldflags "-X main.version=version"
var version string

func main() {

	var (
		configPath      string
		logPath         string
		isLogVerbose    bool
		processorNumber int
	)

	flag.StringVar(&configPath, "c", "./config.json", "config file path")
	flag.StringVar(&logPath, "l", "./overture.log", "log file path")
	flag.BoolVar(&isLogVerbose, "v", false, "verbose mode")
	flag.IntVar(&processorNumber, "p", runtime.NumCPU(), "number of processor to use")
	flag.Parse()

	if isLogVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	logf, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		println("Logfile error: Please check your log file path")
	} else {
		log.SetOutput(io.MultiWriter(logf, os.Stdout))
	}

	log.Info("Overture " + version)
	log.Info("If you need any help, please visit the project repository: https://github.com/shawn1m/overture")

	runtime.GOMAXPROCS(processorNumber)

	core.InitServer(configPath)
}
