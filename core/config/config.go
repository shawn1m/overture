// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package config

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/matcher"
	"github.com/shawn1m/overture/core/matcher/full"
	"github.com/shawn1m/overture/core/matcher/mix"
	"github.com/shawn1m/overture/core/matcher/regex"
	"github.com/shawn1m/overture/core/matcher/suffix"
)

type Config struct {
	BindAddress           string
	DebugHTTPAddress      string
	PrimaryDNS            []*common.DNSUpstream
	AlternativeDNS        []*common.DNSUpstream
	OnlyPrimaryDNS        bool
	IPv6UseAlternativeDNS bool
	IPNetworkFile         struct {
		Primary     string
		Alternative string
	}
	DomainFile struct {
		Primary     string
		Alternative string
		Matcher     string
	}
	HostsFile     string
	MinimumTTL    int
	DomainTTLFile string
	CacheSize     int
	RejectQType   []uint16

	DomainTTLMap                map[string]uint32
	DomainPrimaryList           matcher.Matcher
	DomainAlternativeList       matcher.Matcher
	WhenPrimaryDNSAnswerNoneUse string
	IPNetworkPrimaryList        []*net.IPNet
	IPNetworkAlternativeList    []*net.IPNet
	Hosts                       *hosts.Hosts
	Cache                       *cache.Cache
}

// New config with json file and do some other initiate works
func NewConfig(configFile string) *Config {
	config := parseJson(configFile)

	config.DomainTTLMap = getDomainTTLMap(config.DomainTTLFile)

	config.DomainPrimaryList = initDomainMatcher(config.DomainFile.Primary, config.DomainFile.Matcher)
	config.DomainAlternativeList = initDomainMatcher(config.DomainFile.Alternative, config.DomainFile.Matcher)

	config.IPNetworkPrimaryList = getIPNetworkList(config.IPNetworkFile.Primary)
	config.IPNetworkAlternativeList = getIPNetworkList(config.IPNetworkFile.Alternative)

	if config.MinimumTTL > 0 {
		log.Infof("Minimum TTL has been set to %d", config.MinimumTTL)
	} else {
		log.Info("Minimum TTL is disabled")
	}

	config.Cache = cache.New(config.CacheSize)
	if config.CacheSize > 0 {
		log.Infof("CacheSize is %d", config.CacheSize)
	} else {
		log.Info("Cache is disabled")
	}

	h, err := hosts.New(config.HostsFile)
	if err != nil {
		log.Warnf("Failed to load hosts file: %s", err)
	} else {
		config.Hosts = h
		log.Info("Hosts file has been loaded successfully")
	}

	return config
}

func parseJson(path string) *Config {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %s", err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		os.Exit(1)
	}

	j := new(Config)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s", err)
		os.Exit(1)
	}

	return j
}

func getDomainTTLMap(file string) map[string]uint32 {
	if file == "" {
		return map[string]uint32{}
	}

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read file %s: %s", file, err)
		return nil
	}

	lines := 0
	s := string(f)
	dtl := map[string]uint32{}

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		words := strings.Fields(line)
		tempInt64, err := strconv.ParseUint(words[1], 10, 32)
		dtl[words[0]] = uint32(tempInt64)
		if err != nil {
			log.WithFields(log.Fields{"domain": words[0], "ttl": words[1]}).Warnf("Invalid TTL for domain %s: %s", words[0], words[1])
		}
		lines++
	}

	if len(dtl) > 0 {
		log.Infof("Domain TTL file %s has been loaded with %d records", file, lines)
	} else {
		log.Warnf("There is no element in domain TTL file: %s", file)
	}

	return dtl
}

func getDomainMatcher(name string) (m matcher.Matcher) {
	switch name {
	case "suffix-tree":
		return suffix.DefaultDomainTree()
	case "full-map":
		return &full.Map{DataMap: make(map[string]struct{}, 100)}
	case "full-list":
		return &full.List{}
	case "regex-list":
		return &regex.List{}
	case "mix-list":
		return &mix.List{}
	default:
		log.Warnf("Matcher %s does not exist, using regex-list matcher as default", name)
		return &regex.List{}
	}
}

func initDomainMatcher(file string, name string) (m matcher.Matcher) {
	m = getDomainMatcher(name)

	if file == "" {
		return
	}

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read file %s: %s", file, err)
		return nil
	}

	lines := 0
	s := string(f)

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		_ = m.Insert(line)
		lines++
	}

	if lines > 0 {
		log.Infof("Domain file %s has been loaded with %d records (%s)", file, lines, m.Name())
	} else {
		log.Warnf("There is no element in domain file: %s", file)
	}

	return
}

func getIPNetworkList(file string) []*net.IPNet {
	var ipNetList []*net.IPNet

	// FIXME: why use different file reading mechanism for DomainTTL/Domain and this?
	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open IP network file: %s", err)
		return nil
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		_, ip_net, err := net.ParseCIDR(s.Text())
		if err != nil {
			break
		}
		ipNetList = append(ipNetList, ip_net)
	}

	if len(ipNetList) > 0 {
		log.Infof("IP network file %s has been successfully loaded", file)
	} else {
		log.Warnf("There is no element in IP network file: %s", file)
	}

	return ipNetList
}
