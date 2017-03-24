// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package hosts

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core/common"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type hostname struct {
	domain   string
	ip       net.IP
	ipv6     bool
	wildcard bool
}

type hostnameList []*hostname

func (h *hostname) Equal(he *hostname) bool {
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

// newHostlist creates a hostlist by parsing a file
func newHostnameList(data []byte) *hostnameList {

	ds := string(data)
	hl := new(hostnameList)

	isBar := false
	var bar *pb.ProgressBar
	defer common.TimeTrack(time.Now(), "Load hosts")
	lineList := strings.Split(ds, "\n")
	if len(lineList) >= 10000 {
		isBar = true
	}
	if isBar {
		log.Info("Prepare to load hosts ...")
		bar = pb.StartNew(len(lineList))
		bar.SetRefreshRate(100 * time.Microsecond)
	}

	wg := new(sync.WaitGroup)
	for _, l := range lineList {
		wg.Add(1)
		go func(l string) {
			if isBar {
				defer bar.Increment()
			}
			defer wg.Done()
			if h := parseLine(l); h != nil {
				err := hl.add(h)
				if err != nil {
					log.Warnf("Bad formatted hostsfile line: %s", err)
				}
			}
		}(l)
	}
	wg.Wait()

	if isBar {
		bar.Finish()
	}

	return hl
}

// return first match
func (hl *hostnameList) FindHost(name string) (addr net.IP) {

	var ips []net.IP
	ips = hl.FindHosts(name)
	if len(ips) > 0 {
		addr = ips[0]
	}
	return
}

// return exact matches, if existing -> else, return wildcard
func (hl *hostnameList) FindHosts(name string) (addrs []net.IP) {
	for _, hostname := range *hl {
		if hostname.wildcard == false && hostname.domain == name {
			addrs = append(addrs, hostname.ip)
		}
	}

	if len(addrs) == 0 {
		var domain_match string
		for _, hostname := range *hl {
			if hostname.wildcard == true && len(hostname.domain) < len(name) {
				domain_match = strings.Join([]string{".", hostname.domain}, "")
				if name[len(name)-len(domain_match):] == domain_match {
					var left string
					left = name[0 : len(name)-len(domain_match)]
					if !strings.Contains(left, ".") {
						addrs = append(addrs, hostname.ip)
					}
				}
			}
		}
	}
	return
}

func (hl *hostnameList) add(h *hostname) error {
	for _, found := range *hl {
		if found.Equal(h) {
			return fmt.Errorf("Duplicate hostname entry for %#v", h)
		}
	}
	*hl = append(*hl, h)
	return nil
}

func parseLine(line string) *hostname {

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
	return &hostname{h, ip, isIPv6, isWildcard}
}
