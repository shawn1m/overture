package main

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	setup()
	os.Exit(m.Run())
	shutdown()
}

func setup(){

}

func shutdown(){

}
