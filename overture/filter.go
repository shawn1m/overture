package overture

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"net"
	"strings"
)

func DNSServerFilter(server *dnsServer, message *dns.Msg) {

	log.Debug("Question: " + message.Question[0].String())

	question_name := message.Question[0].Name[:len(message.Question[0].Name)-1]

	if isQuestionInIPv6(message) && Config.RedirectIPv6Record {
		*server = Config.AlternativeDNSServer
		return
	}

	for _, domain := range Config.DomainList {

		if question_name == domain || strings.HasSuffix(question_name, "."+domain) {
			log.Debug("Matched: Custom domain " + question_name + " " + domain)
			*server = Config.AlternativeDNSServer
			return
		}
	}

	log.Debug("Domain match fail, try to use primary DNS.")

	*server =  Config.PrimaryDNSServer
}

func PrimaryDNSResponseFilter(response_message *dns.Msg, question_message *dns.Msg, remote_ip string, ip_net_list []*net.IPNet) {

	if len(response_message.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		err := getResponse(response_message, question_message, remote_ip, Config.AlternativeDNSServer)
		if err != nil {
			log.Warn("Get dns response failed: ", err)
		}
		return
	}

	for _, answer := range response_message.Answer {
		if answer.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network")
		if isIPMatchList(net.ParseIP(answer.(*dns.A).A.String()), ip_net_list, true) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS")
		err := getResponse(response_message, question_message, remote_ip, Config.AlternativeDNSServer)
		if err != nil {
			log.Warn("Get dns response failed: ", err)
		}
		return
	}

	log.Debug("Finally use primary DNS")
}

func EDNSClientSubnetFilter(message *dns.Msg, inbound_ip string) {

	switch Config.EDNSClientSubnetPolicy {
	case "custom":
		setEDNSClientSubnet(message, Config.EDNSClientSubnetIP)
	case "auto":
		if !isIPMatchList(net.ParseIP(inbound_ip), Config.ReservedIPNetworkList, false) {
			setEDNSClientSubnet(message, inbound_ip)
		} else {
			setEDNSClientSubnet(message, Config.ExternalIP)
		}
	case "disable":
	}
}

func MinimumTTLFilter (message *dns.Msg, ttl uint32) {

	if Config.MinimumTTL > 0 {
		setMinimumTTL(message, uint32(ttl))
	}
}

func setMinimumTTL(message *dns.Msg, ttl uint32) {

	for _, answer := range message.Answer {
		if answer.Header().Ttl < ttl {
			answer.Header().Ttl = ttl
		}
	}
}

func setEDNSClientSubnet(message *dns.Msg, ip string) {

	if ip == ""{
		return
	}

	option := message.IsEdns0()

	if option == nil{
		option = new(dns.OPT)
		option.Hdr.Name = "."
		option.Hdr.Rrtype = dns.TypeOPT
		message.Extra = append(message.Extra, option)
	}

	edns0_subnet := isEDNSClientSubnet(option)

	if edns0_subnet == nil{
		edns0_subnet = new(dns.EDNS0_SUBNET)
		option.Option = append(option.Option, edns0_subnet)
	}

	edns0_subnet.Code = dns.EDNS0SUBNET
	edns0_subnet.Address = net.ParseIP(ip)
	if edns0_subnet.Address.To4() != nil {
		edns0_subnet.Family = 1         // 1 for IPv4 source address, 2 for IPv6
		edns0_subnet.SourceNetmask = 32 // 32 for IPV4, 128 for IPv6
	} else {
		edns0_subnet.Family = 2          // 1 for IPv4 source address, 2 for IPv6
		edns0_subnet.SourceNetmask = 128 // 32 for IPV4, 128 for IPv6
	}
	edns0_subnet.SourceScope = 0
}


func isEDNSClientSubnet(option *dns.OPT) *dns.EDNS0_SUBNET{
	for _, s := range option.Option {
		switch e := s.(type) {
		case *dns.EDNS0_SUBNET:
			return e
		}
	}
	return nil
}