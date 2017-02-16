// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package hosts

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type hostlist []*hostname

type hostname struct {
	domain   string
	ip       net.IP
	ipv6     bool
	wildcard bool
}

// newHostlist creates a hostlist by parsing a file
func newHostlist(data []byte) *hostlist {
	return newHostlistString(string(data));
}

func newHostlistString(data string) *hostlist {
	hostlist := hostlist{}
	for _, v := range strings.Split(data, "\n") {
		for _, hostname := range parseLine(v) {
			err := hostlist.add(hostname)
			if err != nil {
				log.Warnf("Bad formatted hostsfile line: %s", err)
			}
		}
	}
	return &hostlist
}

func (h *hostname) Equal(hostnamev *hostname) bool {
	if (h.wildcard != hostnamev.wildcard || h.ipv6 != hostnamev.ipv6) {
		return false
	}
	if (!h.ip.Equal(hostnamev.ip)) {
		return false
	}
	if (h.domain != hostnamev.domain) {
		return false
	}
	return true
}

// return first match
func (h *hostlist) FindHost(name string) (addr net.IP) {
	var ips []net.IP;
	ips = h.FindHosts(name)
	if len(ips) > 0 {
		addr = ips[0];
	}
	return
}

// return exact matches, if existing -> else, return wildcard
func (h *hostlist) FindHosts(name string) (addrs []net.IP) {
	for _, hostname := range *h {
		if hostname.wildcard == false && hostname.domain == name {
			addrs = append(addrs, hostname.ip)
		}
	}

	if len(addrs) == 0 {
		var domain_match string;
		for _, hostname := range *h {
			if hostname.wildcard == true && len(hostname.domain) < len(name) {
				domain_match = strings.Join([]string{".", hostname.domain}, "");
				if name[len(name)-len(domain_match):] == domain_match {
					var left string;
					left = name[0:len(name)-len(domain_match)]
					if !strings.Contains(left, ".") {
						addrs = append(addrs, hostname.ip)
					}
				}
			}
		}
	}

	return
}

func (h *hostlist) add(hostnamev *hostname) error {
	hostname := newHostname(hostnamev.domain, hostnamev.ip, hostnamev.ipv6, hostnamev.wildcard)
	for _, found := range *h {
		if found.Equal(hostname) {
			return fmt.Errorf("Duplicate hostname entry for %#v", hostname)
		}
	}
	*h = append(*h, hostname)
	return nil
}

// newHostname creates a new Hostname struct
func newHostname(domain string, ip net.IP, ipv6 bool, wildcard bool) (host *hostname) {
	domain = strings.ToLower(domain)
	host = &hostname{domain, ip, ipv6, wildcard}
	return
}

// ParseLine parses an individual line in a hostfile, which may contain one
// (un)commented ip and one or more hostnames. For example
//
//	127.0.0.1 localhost mysite1 mysite2
func parseLine(line string) hostlist {
	var hostnames hostlist

	if len(line) == 0 {
		return hostnames
	}

	// Parse leading # for disabled lines
	if line[0:1] == "#" {
		return hostnames
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
	for idx, word := range words {
		words[idx] = strings.TrimSpace(word)
	}

	// Separate the first bit (the ip) from the other bits (the domains)
	address := words[0]
	domains := words[1:]

	if strings.Contains(address, "%") {
		return hostnames
	}

	ip := net.ParseIP(address)

	var isIPv6 bool

	switch {
	case !ip.IsGlobalUnicast() && !ip.IsLoopback():
		return hostnames
	case ip.Equal(net.ParseIP("fe00::")):
		return hostnames
	case ip.To4() != nil:
		isIPv6 = false
	case ip.To16() != nil:
		isIPv6 = true
	default:
		log.Warnf("Invalid IP address found in hostsfile: %s", address)
		return hostnames
	}

	var isWildcard bool
	for _, v := range domains {
		isWildcard = false
		if v[0:2] == "*." {
			v = v[2:]
			isWildcard = true
		}
		hostname := newHostname(v, ip, isIPv6, isWildcard)
		hostnames = append(hostnames, hostname)
	}

	return hostnames
}

// hostsFileMetadata returns metadata about the hosts file.
func hostsFileMetadata(path string) (time.Time, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}, 0, err
	}

	return fi.ModTime(), fi.Size(), nil
}
