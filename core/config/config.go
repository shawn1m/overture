package config

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"

)

var Config *configType

type DNSUpstream struct {
	Address  string
	Protocol string
}

type jsonType struct {
	BindAddress            string `json:"BindAddress"`
	PrimaryDNSAddress      string `json:"PrimaryDNSAddress"`
	PrimaryDNSProtocol     string `json:"PrimaryDNSProtocol"`
	AlternativeDNSAddress  string `json:"AlternativeDNSAddress"`
	AlternativeDNSProtocol string `json:"AlternativeDNSProtocol"`
	Timeout                int    `json:"Timeout"`
	RedirectIPv6Record     bool   `json:"RedirectIPv6Record"`
	IPNetworkFilePath      string `json:"IPNetworkFilePath"`
	DomainFilePath         string `json:"DomainFilePath"`
	DomainBase64Decode     bool   `json:"DomainBase64Decode"`
	MinimumTTL             int    `json:"MinimumTTL"`
	EDNSClientSubnetPolicy string `json:"EDNSClientSubnetPolicy"`
	EDNSClientSubnetIP     string `json:"EDNSClientSubnetIP"`
}

type configType struct {
	BindAddress            string
	PrimaryDNSServer       *DNSUpstream
	AlternativeDNSServer   *DNSUpstream
	Timeout                int
	RedirectIPv6Record     bool
	IPNetworkFilePath      string
	DomainFilePath         string
	DomainBase64Decode     bool
	MinimumTTL             int
	EDNSClientSubnetPolicy string
	EDNSClientSubnetIP     string

	DomainList             []string
	IPNetworkList          []*net.IPNet
	ExternalIP             string
	ReservedIPNetworkList  []*net.IPNet
}

func parseJson(path string) *jsonType {

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

	j := new(jsonType)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatal("Json syntex error: ", err)
		os.Exit(1)
	}

	log.Debug(string(b))

	return j
}

func parseConfig(path string) *configType {

	j := parseJson(path)
	c := &configType{
		BindAddress: j.BindAddress,
		PrimaryDNSServer: &DNSUpstream{
			Address:  j.PrimaryDNSAddress,
			Protocol: j.PrimaryDNSProtocol,
		},
		AlternativeDNSServer: &DNSUpstream{
			Address:  j.AlternativeDNSAddress,
			Protocol: j.AlternativeDNSProtocol,
		},
		Timeout:                j.Timeout,
		RedirectIPv6Record:     j.RedirectIPv6Record,
		IPNetworkFilePath:      j.IPNetworkFilePath,
		DomainFilePath:         j.DomainFilePath,
		DomainBase64Decode:     j.DomainBase64Decode,
		MinimumTTL:             j.MinimumTTL,
		EDNSClientSubnetPolicy: j.EDNSClientSubnetPolicy,
		EDNSClientSubnetIP:     j.EDNSClientSubnetIP,
	}

	return c
}

func NewConfig(path string) *configType{

	c := parseConfig(path)

	return c

}



