// Copyright (c) 2016 holyshawn. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package common

import (
	"net"

	log "github.com/Sirupsen/logrus"
)

func IsIPMatchList(ip net.IP, ipnl []*net.IPNet, isLog bool) bool {

	for _, ip_net := range ipnl {
		if ip_net.Contains(ip) {
			if isLog {
				log.Debug("Matched: IP network " + ip.String() + " " + ip_net.String())
			}
			return true
		}
	}

	return false
}
