// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package core

import (
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/config"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/inbound"
)

func Init(configFilePath string) {

	initConfig(configFilePath)

	inbound.InitServer(config.Config.BindAddress)
}

func initConfig(configFile string) {

	config.Config = config.New(configFile)

	config.Config.IPNetworkList = getIPNetworkList(config.Config.IPNetworkFile)
	config.Config.DomainList = getDomainList(config.Config.DomainFile, config.Config.DomainBase64Decode)

	if config.Config.MinimumTTL > 0 {
		log.Info("Minimum TTL is " + strconv.Itoa(config.Config.MinimumTTL))
	} else {
		log.Info("Minimum TTL is disabled")
	}

	config.Config.CachePool = cache.New(config.Config.CacheSize)

	if config.Config.CacheSize > 0 {
		log.Info("CacheSize is " + strconv.Itoa(config.Config.CacheSize))
	} else {
		log.Info("Cache is disabled")
	}

	err := new(error)
	config.Config.Hosts, *err = hosts.New(config.Config.HostsFile, &hosts.Config{0, false})
	if *err != nil {
		log.Info("Load hosts file failed: ", err)
	} else {
		log.Info("Load hosts file successful")
	}

	config.Config.ReservedIPNetworkList = getReservedIPNetworkList()
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
			log.Error("Decode Custom domain failed:", err)
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
	localCIDR := []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	for _, c := range localCIDR {
		_, ip_net, err := net.ParseCIDR(c)
		if err != nil {
			break
		}
		ipnl = append(ipnl, ip_net)
	}
	return ipnl
}
