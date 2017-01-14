package overture

import (
	"errors"
	"github.com/miekg/dns"
	"reflect"
	"time"
	"net"
)

func getResponse(response_message *dns.Msg, question_message *dns.Msg, remote_address string, dns_server dnsServer) error {

	if reflect.DeepEqual(dns_server, Config.PrimaryDNSServer) {
		switch Config.EDNSClientSubnetPolicy {
		case "custom":
			setEDNSClientSubnet(question_message, Config.EDNSClientSubnetIP)
		case "auto":
			if !isIPMatchList(net.ParseIP(remote_address), Const.ReservedIPNetworkList) {
				setEDNSClientSubnet(question_message, remote_address)
			}else {
				setEDNSClientSubnet(question_message, Const.ExternalIPAddress)
			}
		case "disable":
		}
	}
	client := new(dns.Client)
	client.Net = dns_server.Protocol
	client.Timeout = time.Duration(Config.Timeout) * time.Second
	temp_message, _, err := client.Exchange(question_message, dns_server.Address)
	if temp_message == nil {
		err = errors.New("Response message is nil, maybe timeout.")
		return err
	}
	*response_message = *temp_message
	return err
}
