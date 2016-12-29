package overture

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"net"
	"strings"
)

func chooseDNSAddr(question_data []byte) string {

	message := new(dns.Msg)
	message.Unpack(question_data)

	log.Debug("Question: " + message.Question[0].String())

	question_name := message.Question[0].Name[:len(message.Question[0].Name)-1]

	if isQuestionInIPv6(message) && Config.RedirectIPv6Record {
		return Config.AlternativeDNSAddress
	}

	for _, domain := range custom_domain_list {

		if strings.HasSuffix(question_name, domain) {
			log.Debug("Matched: Custom domain " + question_name + " " + domain)
			return Config.AlternativeDNSAddress
		}
	}

	log.Debug("Domain match fail, try to use primary DNS.")

	return Config.PrimaryDNSAddress
}

func MatchDomesticIPResponse(response_data *[]byte, question_data []byte) {

	message := new(dns.Msg)
	message.Unpack(*response_data)
	for _, answer := range message.Answer {
		if answer.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network.")
		if isIPMatchList(net.ParseIP(answer.(*dns.A).A.String()), ip_net_list) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS.")
		*response_data = getResponse(question_data, Config.AlternativeDNSAddress)
		return
	}

	log.Debug("Finally use primary DNS.")
}

func logResponseAnswer(response_data []byte) {

	message := new(dns.Msg)
	message.Unpack(response_data)
	for _, answer := range message.Answer {
		log.Debug("Answer: " + answer.String())
	}
}

func isIPMatchList(ip net.IP, ip_net_list []*net.IPNet) bool {

	for _, ip_net := range ip_net_list {
		if ip_net.Contains(ip) {
			log.Debug("Matched: IP network " + ip.String() + " " + ip_net.String())
			return true
		}
	}

	return false
}

func isQuestionInIPv6(message *dns.Msg) bool {

	return message.Question[0].Qtype == dns.TypeAAAA
}
