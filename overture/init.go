package overture

import (
	"bufio"
	"encoding/base64"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"github.com/miekg/dns"
	"strconv"
)

var Config *configType

func Init(config_file_path string) {

	Config = parseConfig(config_file_path)

	Config.IPNetworkList = getIPNetworkList(Config.IPNetworkFilePath)
	Config.DomainList = getDomainList(Config.DomainFilePath, Config.DomainBase64Decode)
	switch Config.EDNSClientSubnetPolicy {
	case "auto":
		log.Info("EDNS client subnet auto mode")
		Config.ExternalIPAddress = getExternalIPAddress()
		Config.ReservedIPNetworkList = getReservedIPNetworkList()
	case "custom":
		log.Info("EDNS client subnet custom mode with " + Config.EDNSClientSubnetIP)
	case "disable":
		log.Info("EDNS client subnet disabled")
	}

	if Config.MinimumTTL > 0 {
		log.Info("Minimum TTL is " + strconv.Itoa(Config.MinimumTTL))
	}

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
		log.Info("Load domain file successful")
	} else {
		log.Warn("There is no element in domain file")
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
		log.Info("Load IP network file successful")
	} else {
		log.Warn("There is no element in IP network file")
	}

	return result
}

func getReservedIPNetworkList() []*net.IPNet {

	result := make([]*net.IPNet, 0)
	local_cidr_list := []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	for _, local_cidr := range local_cidr_list {
		_, ip_net, err := net.ParseCIDR(local_cidr)
		if err != nil {
			break
		}
		result = append(result, ip_net)
	}
	return result
}

func getExternalIPAddress() string {
	timeout := time.Duration(Config.Timeout) * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	host := "ip.cn"
	dns_client := dns.Client{}
	dns_message := dns.Msg{}
	dns_message.SetQuestion(host + ".", dns.TypeA)
	response, _, err := dns_client.Exchange(&dns_message, Config.PrimaryDNSServer.Address)
	if err != nil {
		log.Warn("DNS lookup for external ip failed, please check your internet configuration:", err)
		return ""
	}
	request, err := http.NewRequest("GET", "http://" + response.Answer[0].(*dns.A).A.String(), nil)
	if err != nil {
		log.Warn("Get external IP address failed: ", err)
		return ""
	}
	request.Host = host
	resp, err:= client.Do(request)
	if err != nil {
		log.Warn("Get external IP address failed:", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("Get external IP address failed:", err)
		return ""
	}
	re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	result := re.FindString(string(body))
	if len(result) == 0 {
		log.Warn("External IP address is empty")
	}
	log.Info("External IP is " + result)
	return result
}
