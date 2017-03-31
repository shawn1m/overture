// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package core

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"bufio"
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/outbound"
)

type Config struct {
	BindAddress        string `json:"BindAddress"`
	PrimaryDNS         []*outbound.DNSUpstream
	AlternativeDNS     []*outbound.DNSUpstream
	OnlyPrimaryDNS     bool
	RedirectIPv6Record bool
	IPNetworkFile      string
	DomainFile         string
	DomainBase64Decode bool
	HostsFile          string
	MinimumTTL         int
	CacheSize          int
	RejectQtype        []uint16

	DomainList            []string
	IPNetworkList         []*net.IPNet
	Hosts                 *hosts.Hosts
	ReservedIPNetworkList []*net.IPNet
	CachePool             *cache.Cache
}

func NewConfig(configFile string) *Config {

	config := ParseJson(configFile)

	config.IPNetworkList = getIPNetworkList(config.IPNetworkFile)
	config.DomainList = getDomainList(config.DomainFile, config.DomainBase64Decode)
	config.ReservedIPNetworkList = getReservedIPNetworkList()

	if config.MinimumTTL > 0 {
		log.Info("Minimum TTL is " + strconv.Itoa(config.MinimumTTL))
	} else {
		log.Info("Minimum TTL is disabled")
	}

	config.CachePool = cache.New(config.CacheSize)
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

func getDomainList(path string, isBase64 bool) []string {

	var dl []string
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("Open Custom domain file failed: ", err)
		return nil
	}

	re := regexp.MustCompile(`([\w\-\_]+\.[\w\.\-\_]+)[\/\*]*`)
	if isBase64 {
		fd, err := base64.StdEncoding.DecodeString(string(f))
		if err != nil {
			log.Error("Decode Custom domain failed: ", err)
			return nil
		}
		fds := string(fd)
		n := strings.Index(fds, "Whitelist Start")
		dl = re.FindAllString(fds[:n], -1)
	} else {
		dl = re.FindAllString(string(f), -1)
	}

	if len(dl) > 0 {
		log.Info("Load domain file successful")
	} else {
		log.Warn("There is no element in domain file")
	}
	return dl
}

func getIPNetworkList(path string) []*net.IPNet {

	ipnl := make([]*net.IPNet, 0)
	f, err := os.Open(path)
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
		log.Info("Load IP network file successful")
	} else {
		log.Warn("There is no element in IP network file")
	}

	return ipnl
}

func getReservedIPNetworkList() []*net.IPNet {

	ipnl := make([]*net.IPNet, 0)
	localCIDR := []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10"}
	for _, c := range localCIDR {
		_, ip_net, err := net.ParseCIDR(c)
		if err != nil {
			break
		}
		ipnl = append(ipnl, ip_net)
	}
	return ipnl
}

func ParseJson(path string) *Config {

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

	log.Debug(string(b))

	return j
}
