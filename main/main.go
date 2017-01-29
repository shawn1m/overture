package main

import (
	"flag"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core"
)

func main() {

	var (
		configPath      string
		logVerbose      bool
		processorNumber int
	)

	flag.StringVar(&configPath, "c", "./config.json", "config file path")
	flag.BoolVar(&logVerbose, "v", false, "verbose mode")
	flag.IntVar(&processorNumber, "p", runtime.NumCPU(), "number of processor to use")
	flag.Parse()

	if logVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Info("If you need any help or want to check update, please visit the project repository: https://github.com/holyshawn/overture")

	runtime.GOMAXPROCS(processorNumber)

	core.Init(configPath)
}
