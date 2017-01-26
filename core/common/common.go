package common

import (
	"github.com/miekg/dns"
	"net"

	log "github.com/Sirupsen/logrus"
)

func LogAnswer(m *dns.Msg) {

	for _, a := range m.Answer {
		log.Debug("Answer: " + a.String())
	}
}

func IsIPMatchList(ip net.IP, ipnl []*net.IPNet, isLog bool) bool {

	for _, ip_net := range ipnl {
		if ip_net.Contains(ip) {
			if isLog {
				log.Debug("Matched: IP network " + ip.String() + " " + ip_net.String())
			}
			return true
		}
	}

	return false
}

func IsQuestionInIPv6(m *dns.Msg) bool {

	return m.Question[0].Qtype == dns.TypeAAAA
}
