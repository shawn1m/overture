// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package core implements the essential features.
package core

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/shawn1m/overture/core/config"
	"github.com/shawn1m/overture/core/inbound"
	"github.com/shawn1m/overture/core/outbound"
	log "github.com/sirupsen/logrus"
)

var (
	srv      *inbound.Server
	conf     *config.Config
	reloaded bool
)

// Initiate the server with config file
func InitServer(configFilePath string) {
	conf = config.NewConfig(configFilePath)
	go func() {
		StartMonitor()
	}()
	Start()
}

func Start() {
	// New dispatcher without RemoteClientBundle, RemoteClientBundle must be initiated when server is running
	dispatcher := outbound.Dispatcher{
		PrimaryDNS:                  conf.PrimaryDNS,
		AlternativeDNS:              conf.AlternativeDNS,
		OnlyPrimaryDNS:              conf.OnlyPrimaryDNS,
		WhenPrimaryDNSAnswerNoneUse: conf.WhenPrimaryDNSAnswerNoneUse,
		IPNetworkPrimarySet:         conf.IPNetworkPrimarySet,
		IPNetworkAlternativeSet:     conf.IPNetworkAlternativeSet,
		DomainPrimaryList:           conf.DomainPrimaryList,
		DomainAlternativeList:       conf.DomainAlternativeList,

		RedirectIPv6Record:       conf.IPv6UseAlternativeDNS,
		AlternativeDNSConcurrent: conf.AlternativeDNSConcurrent,
		MinimumTTL:               conf.MinimumTTL,
		DomainTTLMap:             conf.DomainTTLMap,

		Hosts: conf.Hosts,
		Cache: conf.Cache,
	}
	dispatcher.Init()

	srv = inbound.NewServer(conf.BindAddress, conf.DebugHTTPAddress, dispatcher, conf.RejectQType)

	go srv.Run()
}

// Stop server
func Stop() {
	srv.Stop()
}

// Reload config and restart server
func Reload() {
	Stop()
	// Have to wait seconds (may be waiting for server shutdown completly) or we will get json parse ERROR. Unknown reason.
	time.Sleep(time.Second)
	conf = config.NewConfig(conf.FilePath)
	reloaded = false
	Start()
}

// Using fsnotify to watch if config file modified.
// It will call Reload() when got a event.
func StartMonitor() {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error open a file watcher: %s", err)
	}

	defer watch.Close()

	done := make(chan bool)

	// Watching config files
	err = watch.Add(conf.FilePath)
	if err != nil {
		log.Fatalf("Error watching config file: %s", conf.FilePath, err)
	}
	err = watch.Add(conf.DomainTTLFile)
	if err != nil {
		log.Fatalf("Error watching %S file: %s", conf.DomainTTLFile, err)
	}
	err = watch.Add(conf.DomainFile.Primary)
	if err != nil {
		log.Fatalf("Error watching %s file: %s", conf.DomainFile.Primary, err)
	}
	err = watch.Add(conf.DomainFile.Alternative)
	if err != nil {
		log.Fatalf("Error watching %s file: %s", conf.DomainFile.Alternative, err)
	}
	err = watch.Add(conf.IPNetworkFile.Primary)
	if err != nil {
		log.Fatalf("Error watching %s file: %s", conf.IPNetworkFile.Primary, err)
	}
	err = watch.Add(conf.IPNetworkFile.Alternative)
	if err != nil {
		log.Fatalf("Error watching %s file: %s", conf.IPNetworkFile.Alternative, err)
	}
	err = watch.Add(conf.HostsFile.HostsFile)
	if err != nil {
		log.Fatalf("Error watching %s file: %s", conf.HostsFile.HostsFile, err)
	}

	go func() {

		// This is a dirty hack to avoid getting multiple same event
		reloaded = false

		for {
			select {
			case event, ok := <-watch.Events:
				if event.Op&fsnotify.Write == fsnotify.Write && ok {
					if !reloaded {
						log.Warnf("%s file changed, reloading.", event.Name)
						reloaded = true
						go Reload()
					}
				}
			case err, ok := <-watch.Errors:
				if !ok {
					log.Fatalf("File watch error: %s", err)
					return
				}
			}
		}

	}()
	<-done
}
