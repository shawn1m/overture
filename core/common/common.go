// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package common provides common functions.
package common

import (
	"net"
	"time"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
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

func TimeTrack(start time.Time, name string) {

	elapsed := time.Since(start)
	log.Debugf("%s took %s", name, elapsed)
}

func IsAnswerEmpty(m *dns.Msg) bool {

	if len(m.Answer) == 0 {
		return true
	}

	return false
}

func HasSubDomain(s string, sub string) bool{

	return strings.HasSuffix(sub, "." + s) || s == sub
}