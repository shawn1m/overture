package overture

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

type dnsServer struct {
	Address  string
	Protocol string
}

type jsonType struct {
	BindAddress            string
	PrimaryDNSAddress      string
	PrimaryDNSProtocol     string
	AlternativeDNSAddress  string
	AlternativeDNSProtocol string
	Timeout                int
	RedirectIPv6Record     bool
	IPNetworkFilePath      string
	DomainFilePath         string
	DomainBase64Decode     bool
	MinimalTTL             int
	EDNSClientSubnetPolicy string
	EDNSClientSubnetIP     string
}

type configType struct {
	BindAddress            string
	PrimaryDNSServer       dnsServer
	AlternativeDNSServer   dnsServer
	Timeout                int
	RedirectIPv6Record     bool
	IPNetworkFilePath      string
	DomainFilePath         string
	DomainBase64Decode     bool
	MinimalTTL             int
	EDNSClientSubnetPolicy string
	EDNSClientSubnetIP     string
}

func parseJson(path string) *jsonType {

	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Open config file failed: ", err)
		os.Exit(1)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("Read config file failed: ", err)
		os.Exit(1)
	}

	result := new(jsonType)
	json_err := json.Unmarshal(data, result)
	if json_err != nil {
		log.Fatal("Json syntex error: ", err)
		os.Exit(1)
	}

	return result
}

func parseConfig(path string) *configType {

	json_result := parseJson(path)
	config := &configType{
		BindAddress: json_result.BindAddress,
		PrimaryDNSServer: dnsServer{
			Address:  json_result.PrimaryDNSAddress,
			Protocol: json_result.PrimaryDNSProtocol,
		},
		AlternativeDNSServer: dnsServer{
			Address:  json_result.AlternativeDNSAddress,
			Protocol: json_result.AlternativeDNSProtocol,
		},
		Timeout:                 json_result.Timeout,
		RedirectIPv6Record:      json_result.RedirectIPv6Record,
		IPNetworkFilePath:       json_result.IPNetworkFilePath,
		DomainFilePath:          json_result.DomainFilePath,
		DomainBase64Decode:      json_result.DomainBase64Decode,
		MinimalTTL:              json_result.MinimalTTL,
		EDNSClientSubnetPolicy:  json_result.EDNSClientSubnetPolicy,
		EDNSClientSubnetIP: json_result.EDNSClientSubnetIP,
	}

	return config
}
