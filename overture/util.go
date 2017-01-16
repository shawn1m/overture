package overture

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"net"
)

func logAnswer(message *dns.Msg) {

	for _, answer := range message.Answer {
		log.Debug("Answer: " + answer.String())
	}
}

func isIPMatchList(ip net.IP, ip_net_list []*net.IPNet, is_log bool) bool {

	for _, ip_net := range ip_net_list {
		if ip_net.Contains(ip) {
			if is_log {
				log.Debug("Matched: IP network " + ip.String() + " " + ip_net.String())
			}
			return true
		}
	}

	return false
}

func isQuestionInIPv6(message *dns.Msg) bool {

	return message.Question[0].Qtype == dns.TypeAAAA
}
