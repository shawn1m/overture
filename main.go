package main

import (
	"./overture"
	"flag"
	log "github.com/Sirupsen/logrus"
	"runtime"
)

func main() {

	var (
		config_file_path string
		log_verbose      bool
		processor_number int
	)

	flag.StringVar(&config_file_path, "c", "./config.json", "config file path")
	flag.BoolVar(&log_verbose, "v", false, "verbose mode")
	flag.IntVar(&processor_number, "p", runtime.NumCPU(), "number of processor to use")
	flag.Parse()

	if log_verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Info("If you need any help or want to check update, please visit the project repository: https://github.com/holyshawn/overture")

	runtime.GOMAXPROCS(processor_number)

	overture.Init(config_file_path)
}
