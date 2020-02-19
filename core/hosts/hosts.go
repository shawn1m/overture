// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package hosts provides address lookups from hosts file.
package hosts

import (
	"bufio"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shawn1m/overture/core/errors"
	"github.com/shawn1m/overture/core/finder"
	log "github.com/sirupsen/logrus"
)

// Hosts represents a file containing hosts_sample
type Hosts struct {
	filePath string
	finder   finder.Finder
}

type hostsLine struct {
	domain string
	ip     net.IP
	isIpv6 bool
}

func New(path string, finder finder.Finder) (*Hosts, error) {
	if path == "" {
		return nil, nil
	}

	h := &Hosts{filePath: path, finder: finder}
	if err := h.initHosts(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Hosts) Find(name string) (ipv4List []net.IP, ipv6List []net.IP) {
	name = strings.TrimSuffix(name, ".")
	hostsLines := h.findHosts(name)
	for _, hostLine := range hostsLines {
		if hostLine.isIpv6 {
			ipv6List = append(ipv6List, hostLine.ip)
		} else {
			ipv4List = append(ipv4List, hostLine.ip)
		}
	}
	return ipv4List, ipv6List
}

func (h *Hosts) initHosts() error {
	f, err := os.Open(h.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer log.Debugf("%s took %s", "Load hosts", time.Since(time.Now()))

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := h.parseLine(scanner.Text()); err != nil {
			log.Warnf("Bad formatted hosts file line: %s", err)
		}
	}
	return nil
}

func (h *Hosts) findHosts(name string) []hostsLine {
	var result []hostsLine
	ips := h.finder.Get(name)
	for _, ipString := range ips {
		ip := net.ParseIP(ipString)
		var isIPv6 bool
		switch {
		case ip.To4() != nil:
			isIPv6 = false
		case ip.To16() != nil:
			isIPv6 = true
		default:
			log.Warnf("Invalid IP address found in hosts file: %s", ip)
			return []hostsLine{}
		}
		result = append(result, hostsLine{
			domain: name,
			ip:     ip,
			isIpv6: isIPv6,
		})
	}
	return result
}

func (h *Hosts) parseLine(line string) error {
	if len(line) == 0 {
		return nil
	}

	// Parse leading # for disabled lines
	if line[0:1] == "#" {
		return nil
	}

	// Parse other #s for actual comments
	line = strings.Split(line, "#")[0]

	// Replace tabs and spaces with single spaces throughout
	line = strings.Replace(line, "\t", " ", -1)
	for strings.Contains(line, "  ") {
		line = strings.Replace(line, "  ", " ", -1)
	}

	line = strings.TrimSpace(line)

	// Break line into words
	words := strings.Split(line, " ")

	if len(words) < 2 {
		log.Warn("Wrong format")
		return &errors.NormalError{Message: "Wrong format"}
	}
	for i, word := range words {
		words[i] = strings.TrimSpace(word)
	}
	// Separate the first bit (the ip) from the other bits (the domains)
	a, host := words[0], words[1]

	ip := net.ParseIP(a)

	err := h.finder.Insert(host, ip.String())
	if err != nil {
		return err
	}
	return nil
}
