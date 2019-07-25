// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package hosts

import (
	"bufio"
	"io"
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

type hostsLines struct {
	data []*hostsLine
	hash map[string]struct{}
}

func newHostsLineList(r io.Reader) *hostsLines {
	resultLines := new(hostsLines)
	resultLines.hash = make(map[string]struct{})

	defer log.Debugf("%s took %s", "Load hosts", time.Since(time.Now()))

	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("Error reading hosts file: %s", err)
			} else {
				log.Debug("Reading hosts file reached EOF")
			}
			break
		}

		if host := parseLine(line); host != nil {
			if err := resultLines.add(host); err != nil {
				log.Warnf("Bad formatted hosts file line: %s", err)
			}
		}
	}

	return resultLines
}

func (hl *hostsLines) FindHosts(name string) (ipv4List []net.IP, ipv6List []net.IP) {
	for _, h := range hl.data {
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
	if _, ok := hl.hash[h.domain]; !ok {
		hl.data = append(hl.data, h)
		hl.hash[h.domain] = struct{}{}
	} else {
		log.Warnf("Duplicate entry for host %s in hosts file, ignored value: %s", h.domain, h.ip.String())
	}
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
		log.Warnf("Invalid IP address found in hosts file: %s", a)
		return nil
	}

	return &hostsLine{h, ip, isIPv6}
}
