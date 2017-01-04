package overture

import (
	"bufio"
	"encoding/base64"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
)

var (
	ip_net_list        []*net.IPNet
	custom_domain_list []string
)

func Init() {

	ip_net_list = getIPNetworkList(Config.IPNetworkFilePath)
	custom_domain_list = getDomainList(Config.DomainFilePath, Config.DomainBase64Decode)

	log.Info("Start overture on " + Config.BindAddress + ".")
	initServer()
}

func getDomainList(path string, base64_decode bool) []string {

	var result []string
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("Open Custom domain file failed: ", err)
		return nil
	}

	re := regexp.MustCompile(`([\w\-\_]+\.[\w\.\-\_]+)[\/\*]*`)
	if base64_decode {
		file_decoded, err := base64.StdEncoding.DecodeString(string(file))
		if err != nil {
			log.Error("Decode Custom domain failed:", err)
			return nil
		}
		file_decoded_string := string(file_decoded)
		n := strings.Index(file_decoded_string, "Whitelist Start")
		result = re.FindAllString(file_decoded_string[:n], -1)
	} else {
		result = re.FindAllString(string(file), -1)
	}

	if len(result) > 0 {
		log.Info("Load domain file successful.")
	} else {
		log.Warn("There is no element in domain file.")
	}
	return result
}

func getIPNetworkList(path string) []*net.IPNet {

	result := make([]*net.IPNet, 0)
	file, err := os.Open(path)
	if err != nil {
		log.Error("Open IP network file failed:", err)
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, ip_net, err := net.ParseCIDR(scanner.Text())
		if err != nil {
			break
		}
		result = append(result, ip_net)
	}
	if len(result) > 0 {
		log.Info("Load IP network file successful.")
	} else {
		log.Warn("There is no element in IP network file.")
	}

	return result
}
