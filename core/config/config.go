package config

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/cache"
)

var Config *configType

type EDNSClientSubnetType struct {
	Policy   string
	CustomIP string
}

type DNSUpstream struct {
	Name             string
	Address          string
	Protocol         string
	Timeout          int
	EDNSClientSubnet EDNSClientSubnetType
}

type configType struct {
	BindAddress        string         `json:"BindAddress"`
	PrimaryDNS         []*DNSUpstream `json:"PrimaryDNS"`
	AlternativeDNS     []*DNSUpstream `json:"AlternativeDNS"`
	RedirectIPv6Record bool           `json:"RedirectIPv6Record"`
	IPNetworkFilePath  string         `json:"IPNetworkFilePath"`
	DomainFilePath     string         `json:"DomainFilePath"`
	DomainBase64Decode bool           `json:"DomainBase64Decode"`
	MinimumTTL         int            `json:"MinimumTTL"`
	CacheSize          int            `json:"CacheSize"`

	DomainList            []string
	IPNetworkList         []*net.IPNet
	ExternalIP            string
	ReservedIPNetworkList []*net.IPNet
	CachePool             *cache.Cache
}

func parseJson(path string) *configType {

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

	j := new(configType)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatal("Json syntex error: ", err)
		os.Exit(1)
	}

	log.Debug(string(b))

	return j
}

func NewConfig(path string) *configType {

	return parseJson(path)

}
