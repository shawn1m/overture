package core

import (
	"io/ioutil"
	"regexp"
	"time"
	"net"
	"net/http"
	"encoding/base64"
	"strings"
	"os"
	"strconv"
	"bufio"

	"github.com/miekg/dns"
	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/inbound"
	"github.com/holyshawn/overture/core/outbound"
)

func Init(config_file_path string) {

	config.Config = config.NewConfig(config_file_path)

	config.Config.IPNetworkList = getIPNetworkList(config.Config.IPNetworkFilePath)
	config.Config.DomainList = getDomainList(config.Config.DomainFilePath, config.Config.DomainBase64Decode)
	switch config.Config.EDNSClientSubnetPolicy {
	case "auto":
		log.Info("EDNS client subnet auto mode")
		config.Config.ExternalIP = getExternalIP()
		config.Config.ReservedIPNetworkList = getReservedIPNetworkList()
	case "custom":
		log.Info("EDNS client subnet custom mode with " + config.Config.EDNSClientSubnetIP)
	case "disable":
		log.Info("EDNS client subnet disabled")
	}

	if config.Config.MinimumTTL > 0 {
		log.Info("Minimum TTL is " + strconv.Itoa(config.Config.MinimumTTL))
	}

	inbound.InitServer(config.Config.BindAddress)
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
		log.Error("Open IP network file failed:", err)
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

func getExternalIP() string {

	timeout := time.Duration(config.Config.Timeout) * time.Second * 2
	c := http.Client{
		Timeout: timeout,
	}
	host := "ip.cn"
	q := new(dns.Msg)
	q.SetQuestion(host+".", dns.TypeA)
	o := outbound.NewOutbound(q, "", config.Config.PrimaryDNSServer)
	err := o.ExchangeFromRemote()
	if err != nil{
		log.Error("Get external IP address failed, please check your primary DNS ", err)
		return ""
	}
	req, err := http.NewRequest("GET", "http://"+o.ResponseMessage.Answer[0].(*dns.A).A.String(), nil)
	if err != nil {
		log.Warn("Get external IP address failed: ", err)
		return ""
	}
	req.Host = host
	res, err := c.Do(req)
	if err != nil {
		log.Warn("Get external IP address failed: ", err)
		return ""
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warn("Get external IP address failed: ", err)
		return ""
	}
	re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	eip := re.FindString(string(body))
	if len(eip) == 0 {
		log.Warn("External IP address is empty")
	}
	log.Info("External IP is " + eip)
	return eip
}