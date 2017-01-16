package overture

import (
	"errors"
	"github.com/miekg/dns"
	"reflect"
	"time"
)

func getResponse(response_message *dns.Msg, question_message *dns.Msg, inbound_ip string, dns_server dnsServer) error {

	if reflect.DeepEqual(dns_server, Config.PrimaryDNSServer) {
		EDNSClientSubnetFilter(question_message, inbound_ip)
	}

	client := new(dns.Client)
	client.Net = dns_server.Protocol
	client.Timeout = time.Duration(Config.Timeout) * time.Second
	temp_message, _, err := client.Exchange(question_message, dns_server.Address)
	if temp_message == nil {
		err = errors.New("Response message is nil, maybe timeout")
		return err
	}
	*response_message = *temp_message
	return err
}
