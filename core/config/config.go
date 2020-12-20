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

	"github.com/shawn1m/overture/core/finder"
	finderfull "github.com/shawn1m/overture/core/finder/full"
	finderregex "github.com/shawn1m/overture/core/finder/regex"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/matcher"
	matcherfinal "github.com/shawn1m/overture/core/matcher/final"
	matcherfull "github.com/shawn1m/overture/core/matcher/full"
	matchermix "github.com/shawn1m/overture/core/matcher/mix"
	matcherregex "github.com/shawn1m/overture/core/matcher/regex"
	matchersuffix "github.com/shawn1m/overture/core/matcher/suffix"
)

type Config struct {
	FilePath                    string
	BindAddress                 string                `yaml:"bindAddress"`
	DebugHTTPAddress            string                `yaml:"debugHTTPAddress"`
	DohEnabled                  bool                  `yaml:"dohEnabled"`
	PrimaryDNS                  []*common.DNSUpstream `yaml:"primaryDNS"`
	AlternativeDNS              []*common.DNSUpstream `yaml:"alternativeDNS"`
	OnlyPrimaryDNS              bool                  `yaml:"onlyPrimaryDNS"`
	IPv6UseAlternativeDNS       bool                  `yaml:"ipv6UseAlternativeDNS"`
	AlternativeDNSConcurrent    bool                  `yaml:"alternativeDNSConcurrent"`
	WhenPrimaryDNSAnswerNoneUse string                `yaml:"whenPrimaryDNSAnswerNoneUse"`
	IPNetworkFile               struct {
		Primary     string `yaml:"primary"`
		Alternative string `yaml:"alternative"`
	} `yaml:"ipNetworkFile"`
	DomainFile struct {
		Primary            string `yaml:"primary"`
		Alternative        string `yaml:"alternative"`
		PrimaryMatcher     string `yaml:"primaryMatcher"`
		AlternativeMatcher string `yaml:"alternativeMatcher"`
		Matcher            string `yaml:"matcher"`
	} `yaml:"domainFile"`
	HostsFile struct {
		HostsFile string `yaml:"hostsFile"`
		Finder    string `yaml:"finder"`
	} `yaml:"hostsFile"`
	MinimumTTL    int      `yaml:"minimumTTL"`
	DomainTTLFile string   `yaml:"domainTTLFile"`
	CacheSize     int      `yaml:"cacheSize"`
	RejectQType   []uint16 `yaml:"rejectQType"`

	DomainTTLMap            map[string]uint32
	DomainPrimaryList       matcher.Matcher
	DomainAlternativeList   matcher.Matcher
	IPNetworkPrimarySet     *common.IPSet
	IPNetworkAlternativeSet *common.IPSet
	Hosts                   *hosts.Hosts
	Cache                   *cache.Cache
}

// New config with json file and do some other initiate works
func NewConfig(configFile string) *Config {
	config := parseConfigFile(configFile)
	config.FilePath = configFile

	config.DomainTTLMap = getDomainTTLMap(config.DomainTTLFile)

	config.DomainPrimaryList = initDomainMatcher(config.DomainFile.Primary, config.DomainFile.PrimaryMatcher, config.DomainFile.Matcher)
	config.DomainAlternativeList = initDomainMatcher(config.DomainFile.Alternative, config.DomainFile.AlternativeMatcher, config.DomainFile.Matcher)

	config.IPNetworkPrimarySet = getIPNetworkSet(config.IPNetworkFile.Primary)
	config.IPNetworkAlternativeSet = getIPNetworkSet(config.IPNetworkFile.Alternative)

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

	h, err := hosts.New(config.HostsFile.HostsFile, getFinder(config.HostsFile.Finder))
	if err != nil {
		log.Warnf("Failed to load hosts file: %s", err)
	} else {
		config.Hosts = h
		log.Info("Hosts file has been loaded successfully")
	}

	return config
}

func parseConfigFile(path string) *Config {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
		os.Exit(1)
	}

	config := new(Config)
	if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {
		err = yaml.Unmarshal(b, config)
	} else {
		err = json.Unmarshal(b, config)
	}

	if err != nil {
		log.Fatalf("Failed to parse config file: %s", err)
		os.Exit(1)
	}

	return config
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

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
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
		return matchersuffix.DefaultDomainTree()
	case "full-map":
		return &matcherfull.Map{DataMap: make(map[string]struct{}, 100)}
	case "full-list":
		return &matcherfull.List{}
	case "regex-list":
		return &matcherregex.List{}
	case "mix-list":
		return &matchermix.List{}
	case "final":
		return &matcherfinal.Default{}
	default:
		log.Warnf("Matcher %s does not exist, using full-map matcher as default", name)
		return &matcherfull.Map{DataMap: make(map[string]struct{}, 100)}
	}
}

func getFinder(name string) (f finder.Finder) {
	switch name {
	case "regex-list":
		return &finderregex.List{RegexMap: make(map[string][]string, 100)}
	case "full-map":
		return &finderfull.Map{DataMap: make(map[string][]string, 100)}
	default:
		log.Warnf("Finder %s does not exist, using full-map finder as default", name)
		return &finderfull.Map{DataMap: make(map[string][]string, 100)}
	}
}

func initDomainMatcher(file string, name string, defaultName string) (m matcher.Matcher) {
	if name == "" {
		name = defaultName
	}
	m = getDomainMatcher(name)
	if name == "final" {
		return m
	}
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
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
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

func getIPNetworkSet(file string) *common.IPSet {
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

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
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

	return common.NewIPSet(ipNetList)
}
