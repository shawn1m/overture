// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package config

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
)

type Config struct {
	BindAddress        string `json:"BindAddress"`
	PrimaryDNS         []*common.DNSUpstream
	AlternativeDNS     []*common.DNSUpstream
	OnlyPrimaryDNS     bool
	RedirectIPv6Record bool
	IPNetworkFile      string
	DomainFile         string
	DomainWhiteFile    string
	DomainBase64Decode bool
	HostsFile          string
	MinimumTTL         int
	CacheSize          int
	RejectQtype        []uint16

	DomainList      []string
	DomainWhiteList []string
	IPNetworkList   []*net.IPNet
	Hosts           *hosts.Hosts
	Cache           *cache.Cache
}

// New config with json file and do some other initiate works
func NewConfig(configFile string) *Config {

	config := parseJson(configFile)

	config.getIPNetworkList()
	config.getDomainList()
	config.getDomainWhiteList()

	if config.MinimumTTL > 0 {
		log.Info("Minimum TTL is " + strconv.Itoa(config.MinimumTTL))
	} else {
		log.Info("Minimum TTL is disabled")
	}

	config.Cache = cache.New(config.CacheSize)
	if config.CacheSize > 0 {
		log.Info("CacheSize is " + strconv.Itoa(config.CacheSize))
	} else {
		log.Info("Cache is disabled")
	}

	h, err := hosts.New(config.HostsFile)
	if err != nil {
		log.Info("Load hosts file failed: ", err)
	} else {
		config.Hosts = h
		log.Info("Load hosts file successful")
	}

	return config
}

func parseJson(path string) *Config {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Open config file failed: ", err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("Read config file failed: ", err)
		os.Exit(1)
	}

	j := new(Config)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatal("Json syntex error: ", err)
		os.Exit(1)
	}

	return j
}

func (c *Config) getDomainWhiteList() {
	var dl []string
	f, err := ioutil.ReadFile(c.DomainWhiteFile)
	if err != nil {
		log.Error("Open Domain WhiteList file failed: ", err)
		return
	}

	lines := 0
	s := string(f)
	dl = []string{}

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		dl = append(dl, line)
		lines++
	}

	if len(dl) > 0 {
		log.Infof("Load domain whitelist file successful with %d records ", lines)
	} else {
		log.Warn("There is no element in domain whitelist file")
	}
	c.DomainWhiteList = dl
}

func (c *Config) getDomainList() {

	var dl []string
	f, err := ioutil.ReadFile(c.DomainFile)
	if err != nil {
		log.Error("Open Custom domain file failed: ", err)
		return
	}

	re := regexp.MustCompile(`([\w\-\_]+\.[\w\.\-\_]+)[\/\*]*`)
	if c.DomainBase64Decode {
		fd, err := base64.StdEncoding.DecodeString(string(f))
		if err != nil {
			log.Error("Decode Custom domain failed: ", err)
			return
		}
		fds := string(fd)
		n := strings.Index(fds, "Whitelist Start")
		dl = re.FindAllString(fds[:n], -1)
	} else {
		dl = re.FindAllString(string(f), -1)
	}

	uniqueDl := map[string]bool{}
	for _, e := range dl {
		uniqueDl[e] = true
	}

	dl = []string{}

	for k := range uniqueDl {
		dl = append(dl, k)
	}

	if len(dl) > 0 {
		log.Info("Load domain file successful")
	} else {
		log.Warn("There is no element in domain file")
	}
	c.DomainList = dl
}

func (c *Config) getIPNetworkList() {

	ipnl := make([]*net.IPNet, 0)
	f, err := os.Open(c.IPNetworkFile)
	if err != nil {
		log.Error("Open IP network file failed: ", err)
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		_, ip_net, err := net.ParseCIDR(s.Text())
		if err != nil {
			break
		}
		ipnl = append(ipnl, ip_net)
	}
	if len(ipnl) > 0 {
		log.Info("Load IP network file successful")
	} else {
		log.Warn("There is no element in IP network file")
	}

	c.IPNetworkList = ipnl
}
