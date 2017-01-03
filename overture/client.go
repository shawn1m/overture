package overture

import (
	"github.com/miekg/dns"
)

func getResponse(method string, message *dns.Msg, address string) (*dns.Msg) {

	clientTCP := new(dns.Client)
	clientTCP.Net = method
	m, _, _ := clientTCP.Exchange(message, address)
	return m
}