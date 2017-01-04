package overture

import (
	"github.com/miekg/dns"
	"time"
	"errors"
)

func getResponse(response_message *dns.Msg, question_message *dns.Msg, dns_server dnsServer) error {

	client := new(dns.Client)
	client.Net = dns_server.Method
	client.Timeout = time.Duration(Config.Timeout) * time.Second
	temp_message, _, err := client.Exchange(question_message, dns_server.Address)
	if temp_message == nil{
		err = errors.New("Response message is nil, maybe timeout.")
		return err
	}
	*response_message = *temp_message
	return err
}