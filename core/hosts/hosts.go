// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package hosts provides address lookups from hosts file.
package hosts

import (
	"io/ioutil"
	"net"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

// Hosts represents a file containing hosts_sample
type Hosts struct {
	sync.RWMutex
	hl       *hostsLineList
	filePath string
}

func New(path string) (*Hosts, error) {

	if path == "" {
		return nil, nil
	}

	h := &Hosts{filePath: path}
	if err := h.loadHostEntries(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Hosts) Find(name string) []net.IP {
	name = strings.TrimSuffix(name, ".")
	h.RLock()
	defer h.RUnlock()
	return h.hl.FindHosts(name)
}

func (h *Hosts) FindReverse(ip string) string {
	h.RLock()
	defer h.RUnlock()

	for _, hostname := range *h.hl {
		if r, _ := dns.ReverseAddr(hostname.ip.String()); ip == r {
			return dns.Fqdn(hostname.domain)
		}
	}

	return ""
}

func (h *Hosts) loadHostEntries() error {
	data, err := ioutil.ReadFile(h.filePath)
	if err != nil {
		return err
	}

	h.hl = newHostsLineList(data)

	return nil
}
