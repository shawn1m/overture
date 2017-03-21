// Copyright (c) 2015 Jan Broer. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package hosts_sample provides address lookups from local hosts_sample (usually /etc/hosts_sample).
package hosts

import (
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// Config stores options for hosts_sample
type Config struct {
	// Positive value enables polling
	Poll    int
	Verbose bool
}

// Hosts represents a file containing hosts_sample
type Hosts struct {
	config *Config
	hosts  *hostlist
	file   struct {
		size  int64
		path  string
		mtime time.Time
	}
	hostMutex sync.RWMutex
}

// New returns a new Hosts object
func New(path string, config *Config) (*Hosts, error) {
	h := Hosts{config: config}
	// when no hosts_sample file is given we return an empty hostlist
	if path == "" {
		h.hosts = new(hostlist)
		return &h, nil
	}

	h.file.path = path
	if err := h.loadHostEntries(); err != nil {
		return nil, err
	}

	return &h, nil
}

func (h *Hosts) FindHosts(name string) (addrs []net.IP, err error) {
	name = strings.TrimSuffix(name, ".")
	h.hostMutex.RLock()
	defer h.hostMutex.RUnlock()
	addrs = h.hosts.FindHosts(name)
	return
}

func (h *Hosts) FindReverse(name string) (host string, err error) {
	h.hostMutex.RLock()
	defer h.hostMutex.RUnlock()

	for _, hostname := range *h.hosts {
		if r, _ := dns.ReverseAddr(hostname.ip.String()); name == r {
			host = dns.Fqdn(hostname.domain)
			break
		}
	}
	return
}

func (h *Hosts) loadHostEntries() error {
	data, err := ioutil.ReadFile(h.file.path)
	if err != nil {
		return err
	}

	h.hosts = newHostlist(data)

	return nil
}
