// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package hosts

import (
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/common"
)

type hostsLine struct {
	domain string
	ip     net.IP
	ipv6   bool
}

type hostsLines []*hostsLine

func newHostsLineList(data []byte) *hostsLines {

	resultLines := new(hostsLines)

	defer log.Debugf("%s took %s", "Load hosts", time.Since(time.Now()))
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		func(l string) {
			if h := parseLine(l); h != nil {
				err := resultLines.add(h)
				if err != nil {
					log.Warnf("Bad formatted hosts file line: %s", err)
				}
			}
		}(line)
	}

	return resultLines
}

func (hl *hostsLines) FindHosts(name string) (ipv4List []net.IP, ipv6List []net.IP) {

	for _, h := range *hl {
		if common.IsDomainMatchRule(h.domain, name) {
			log.WithFields(log.Fields{
				"question": name,
				"domain":   h.domain,
				"ip":       h.ip,
			}).Debug("Matched")
			if h.ip.To4() != nil {
				ipv4List = append(ipv4List, h.ip)
			} else {
				ipv6List = append(ipv6List, h.ip)
			}
		}
	}
	return
}

func (hl *hostsLines) add(h *hostsLine) error {
	// Use too much CPU time when hosts file is big
	// for _, found := range *hl {
	// 	if found.Equal(h) {
	// 		return fmt.Errorf("Duplicate hostname entry for %#v", h)
	// 	}
	// }
	*hl = append(*hl, h)
	return nil
}

func parseLine(line string) *hostsLine {

	if len(line) == 0 {
		return nil
	}

	// Parse leading # for disabled lines
	if line[0:1] == "#" {
		return nil
	}

	// Parse other #s for actual comments
	line = strings.Split(line, "#")[0]

	// Replace tabs and multispaces with single spaces throughout
	line = strings.Replace(line, "\t", " ", -1)
	for strings.Contains(line, "  ") {
		line = strings.Replace(line, "  ", " ", -1)
	}

	line = strings.TrimSpace(line)

	// Break line into words
	words := strings.Split(line, " ")

	if len(words) < 2 {
		return nil
	}
	for i, word := range words {
		words[i] = strings.TrimSpace(word)
	}
	// Separate the first bit (the ip) from the other bits (the domains)
	a, h := words[0], words[1]

	ip := net.ParseIP(a)

	var isIPv6 bool

	switch {
	case ip.To4() != nil:
		isIPv6 = false
	case ip.To16() != nil:
		isIPv6 = true
	default:
		log.Warnf("Invalid IP address found in hostsfile: %s", a)
		return nil
	}

	return &hostsLine{h, ip, isIPv6}
}
