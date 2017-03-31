// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package core implements the most essential function
package core

import (
	"github.com/shawn1m/overture/core/inbound"
	"github.com/shawn1m/overture/core/outbound"
)

func Init(configFilePath string) {

	config := NewConfig(configFilePath)

	d := &outbound.Dispatcher{
		PrimaryDNS:         config.PrimaryDNS,
		AlternativeDNS:     config.AlternativeDNS,
		IPNetworkList:      config.IPNetworkList,
		DomainList:         config.DomainList,
		RedirectIPv6Record: config.RedirectIPv6Record,
	}

	s := &inbound.Server{
		BindAddress: config.BindAddress,
		Dispatcher: d,
		OnlyPrimaryDNS: config.OnlyPrimaryDNS,
		MinimumTTL: config.MinimumTTL,
		RejectQtype: config.RejectQtype,
		Hosts: config.Hosts,
		CachePool: config.CachePool,
	}

	s.Run()
}
