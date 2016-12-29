package overture

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

var Config configType

type configType struct {
	BindAddress           string
	PrimaryDNSAddress     string
	AlternativeDNSAddress string
	RedirectIPv6Record    bool
	IPNetworkFilePath     string
	DomainFilePath        string
	DomainBase64Decode    bool
}

func ParseConfig(path string) configType {

	file, err := os.Open(path)
	if err != nil {
		log.Error("Open config file failed: ", err)
		os.Exit(1)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error("Read config file failed: ", err)
		os.Exit(1)
	}

	var config configType
	json_err := json.Unmarshal(data, &config)
	if json_err != nil {
		log.Error("Json syntex error: ", err)
		os.Exit(1)
	}

	return config
}
