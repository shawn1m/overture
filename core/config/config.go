// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package config

import (
	"bufio"
	"encoding/json"
	"io"
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
	AlternativeDNSConcurrent bool
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
	b, err := ioutil.ReadFile(path)
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

	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open domain TTL file %s: %s", file, err)
		return nil
	}
	defer f.Close()

	successes := 0
	failures := 0
	var failedLines []string

	dtl := map[string]uint32{}

	reader := bufio.NewReader(f)

	for {
		// The last line may not contains an '\n'
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Errorf("Failed to read domain TTL file %s: %s", file, err)
			break
		}

		if line != "" {
			words := strings.Fields(line)
			if len(words) > 1 {
				tempInt64, err := strconv.ParseUint(words[1], 10, 32)
				dtl[words[0]] = uint32(tempInt64)
				if err != nil {
					log.WithFields(log.Fields{"domain": words[0], "ttl": words[1]}).Warnf("Invalid TTL for domain %s: %s", words[0], words[1])
					failures++
					failedLines = append(failedLines, line)
				}
				successes++
			} else {
				failedLines = append(failedLines, line)
				failures++
			}
		}
		if line == "" && err == io.EOF {
			log.Debugf("Reading domain TTL file %s reached EOF", file)
			break
		}
	}

	if len(dtl) > 0 {
		log.Infof("Domain TTL file %s has been loaded with %d records (%d failed)", file, successes, failures)
		if len(failedLines) > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	} else {
		log.Warnf("No element has been loaded from domain TTL file: %s", file)
		if len(failedLines) > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
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

	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open domain file %s: %s", file, err)
		return nil
	}
	defer f.Close()

	lines := 0
	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Errorf("Failed to read domain file %s: %s", file, err)
			break
		}

		line = strings.TrimSpace(line)
		if line != "" {
			_ = m.Insert(line)
			lines++
		}
		if line == "" && err == io.EOF {
			log.Debugf("Reading domain file %s reached EOF", file)
			break
		}
	}

	if lines > 0 {
		log.Infof("Domain file %s has been loaded with %d records (%s)", file, lines, m.Name())
	} else {
		log.Warnf("No element has been loaded from domain file: %s", file)
	}

	return
}

func getIPNetworkList(file string) []*net.IPNet {
	ipNetList := make([]*net.IPNet, 0)

	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Failed to open IP network file: %s", err)
		return nil
	}
	defer f.Close()

	successes := 0
	failures := 0
	var failedLines []string

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("Failed to read IP network file %s: %s", file, err)
			} else {
				log.Debugf("Reading IP network file %s has reached EOF", file)
			}
			break
		}

		if line != "" {
			_, ipNet, err := net.ParseCIDR(strings.TrimSuffix(line, "\n"))
			if err != nil {
				log.Errorf("Error parsing IP network CIDR %s: %s", line, err)
				failures++
				failedLines = append(failedLines, line)
				continue
			}
			ipNetList = append(ipNetList, ipNet)
			successes++
		}
	}

	if len(ipNetList) > 0 {
		log.Infof("IP network file %s has been loaded with %d records", file, successes)
		if failures > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	} else {
		log.Warnf("No element has been loaded from IP network file: %s", file)
		if failures > 0 {
			log.Debugf("Failed lines (%s):", file)
			for _, line := range failedLines {
				log.Debug(line)
			}
		}
	}

	return ipNetList
}
