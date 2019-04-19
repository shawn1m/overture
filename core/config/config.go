// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package config

import (
	"bufio"
	"encoding/json"

	"github.com/shawn1m/overture/core/matcher"
	"github.com/shawn1m/overture/core/matcher/full"
	"github.com/shawn1m/overture/core/matcher/regex"
	"github.com/shawn1m/overture/core/matcher/suffix"
	"github.com/shawn1m/overture/core/matcher/mix"

	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
)

type Config struct {
	BindAddress           string `json:"BindAddress"`
	DebugHTTPAddress      string `json:"DebugHTTPAddress"`
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

func getDomainTTLMap(file string) map[string]uint32 {

	if file == "" {
		return map[string]uint32{}
	}

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("Open file "+file+" failed: ", err)
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
			log.WithFields(log.Fields{"domain": words[0], "ttl": words[1]}).Warn("This TTL is not a number!")
		}
		lines++
	}

	if len(dtl) > 0 {
		log.Infof("Load domain TTL "+file+" successful with %d records ", lines)
	} else {
		log.Warn("There is no element in domain TTL file")
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
		return &full.List{DataList: []string{}}
	case "regex-list":
		return &regex.List{RegexList: []string{}}
	case "mix-list":
		return &mix.List{DataList: make([]mix.Data, 0)}
	default:
		log.Warn("There is no such matcher: "+name, ", use regex-list matcher as default")
		return &regex.List{RegexList: []string{}}
	}
}

func initDomainMatcher(file string, name string) (m matcher.Matcher) {

	m = getDomainMatcher(name)

	if file == "" {
		return
	}

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("Open file "+file+" failed: ", err)
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
		log.Infof("Load domain "+file+" successful with %d records ("+m.Name()+")", lines)
	} else {
		log.Warn("There is no element in this domain file: " + file)
	}

	return
}

func getIPNetworkList(file string) []*net.IPNet {

	ipnl := make([]*net.IPNet, 0)
	f, err := os.Open(file)
	if err != nil {
		log.Error("Open IP network file failed: ", err)
		return nil
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
		log.Info("Load " + file + " successful")
	} else {
		log.Warn("There is no element in " + file)
	}

	return ipnl
}
