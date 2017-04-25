// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package hosts

import (
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core/common"
)

type hostsLine struct {
	domain   string
	ip       net.IP
	ipv6     bool
	wildcard bool
}

type hostsLineList []*hostsLine

func (h *hostsLine) Equal(he *hostsLine) bool {
	if h.wildcard != he.wildcard || h.ipv6 != he.ipv6 {
		return false
	}
	if !h.ip.Equal(he.ip) {
		return false
	}
	if h.domain != he.domain {
		return false
	}
	return true
}

func newHostsLineList(data []byte) *hostsLineList {

	ds := string(data)
	hl := new(hostsLineList)

	defer common.TimeTrack(time.Now(), "Load hosts")
	lineList := strings.Split(ds, "\n")

	for _, l := range lineList {
		func(l string) {
			if h := parseLine(l); h != nil {
				err := hl.add(h)
				if err != nil {
					log.Warnf("Bad formatted hostsfile line: %s", err)
				}
			}
		}(l)
	}

	return hl
}

func (hl *hostsLineList) FindHosts(name string) (ipv4List []net.IP, ipv6List []net.IP) {

	for _, h := range *hl {
		if (h.wildcard == false && h.domain == name) ||
			(h.wildcard == true && common.HasSubDomain(h.domain, name)) {
			if h.ip.To4() != nil {
				ipv4List = append(ipv4List, h.ip)
			} else {
				ipv6List = append(ipv6List, h.ip)
			}
		}
	}
	return
}

func (hl *hostsLineList) add(h *hostsLine) error {
	// Use too much CPU time when hosts file is big
	//for _, found := range *hl {
	//	if found.Equal(h) {
	//		return fmt.Errorf("Duplicate hostname entry for %#v", h)
	//	}
	//}
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

	isWildcard := false
	if h[0:2] == "*." {
		h = h[2:]
		isWildcard = true
	}
	return &hostsLine{h, ip, isIPv6, isWildcard}
}
