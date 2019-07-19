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
var (
	version string

	configPath      = flag.String("c", "./config.json", "config file path")
	logPath         = flag.String("l", "", "log file path")
	isLogVerbose    = flag.Bool("v", false, "verbose mode")
	processorNumber = flag.Int("p", runtime.NumCPU(), "number of processor to use")
	isShowVersion   = flag.Bool("V", false, "current version of overture")
)

func main() {
	flag.Parse()

	if *isShowVersion {
		fmt.Println(version)
		return
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if *isLogVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if *logPath != "" {
		lf, err := os.OpenFile(*logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			log.Errorf("Unable to open log file for writing: %s", err)
		} else {
			log.SetOutput(io.MultiWriter(lf, os.Stdout))
		}
	}

	log.Infof("Overture %s", version)
	log.Info("If you need any help, please visit the project repository: https://github.com/shawn1m/overture")

	runtime.GOMAXPROCS(*processorNumber)

	core.InitServer(*configPath)
}
