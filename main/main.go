// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package main is the entry point of whole program.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core"
)

// For auto version building
//  go build -ldflags "-X main.version=version"
var version = ""

func main() {

	var (
		configPath      string
		logPath         string
		isLogVerbose    bool
		processorNumber int
		isShowVersion   bool
	)

	flag.StringVar(&configPath, "c", "./config.json", "config file path")
	flag.StringVar(&logPath, "l", "", "log file path")
	flag.BoolVar(&isLogVerbose, "v", false, "verbose mode")
	flag.IntVar(&processorNumber, "p", runtime.NumCPU(), "number of processor to use")
	flag.BoolVar(&isShowVersion, "V", false, "current version of overture")
	flag.Parse()

	if isShowVersion {
		fmt.Println(version)
		return
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if isLogVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if logPath != "" {
		lf, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			println("Logfile error: Please check your log file path")
		} else {
			log.SetOutput(io.MultiWriter(lf, os.Stdout))
		}
	}

	log.Info("Overture " + version)
	log.Info("If you need any help, please visit the project repository: https://github.com/shawn1m/overture")

	runtime.GOMAXPROCS(processorNumber)

	core.InitServer(configPath)
}
