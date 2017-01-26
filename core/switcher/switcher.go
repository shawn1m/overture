package switcher

import (
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/holyshawn/overture/core/outbound"
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/common"
)

type Switcher struct {
	outbound               *outbound.Outbound
	ipNetworkList          []*net.IPNet
	domainList             []string
	redirectIPv6Record     bool
}

func NewSwitcher(outbound *outbound.Outbound) *Switcher {
	return &Switcher{
		outbound: outbound,
		ipNetworkList: config.Config.IPNetworkList,
		domainList: config.Config.DomainList,
		redirectIPv6Record: config.Config.RedirectIPv6Record,
	}
}

func (s *Switcher) ChooseNameSever() {

	log.Debug("Question: " + s.outbound.QuestionMessage.Question[0].String())

	qn := s.outbound.QuestionMessage.Question[0].Name[:len(s.outbound.QuestionMessage.Question[0].Name)-1]

	if common.IsQuestionInIPv6(s.outbound.QuestionMessage) && s.redirectIPv6Record {
		s.outbound.DNSUpstream = config.Config.AlternativeDNSServer
		return
	}

	for _, d := range s.domainList {

		if qn == d || strings.HasSuffix(qn, "."+ d) {
			log.Debug("Matched: Custom domain " + qn + " " + d)
			s.outbound.DNSUpstream = config.Config.AlternativeDNSServer
			return
		}
	}

	log.Debug("Domain match fail, try to use primary DNS")

	s.outbound.DNSUpstream = config.Config.PrimaryDNSServer
}

func (s *Switcher) HandleResponseFromPrimaryDNS() {

	if len(s.outbound.ResponseMessage.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		s.outbound.DNSUpstream = config.Config.AlternativeDNSServer
		err := s.outbound.ExchangeFromRemote()
		if err != nil {
			log.Warn("Get dns response failed: ", err)
		}
		return
	}

	for _, a := range s.outbound.ResponseMessage.Answer {
		if a.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network")
		if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), s.ipNetworkList, true) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS")
		s.outbound.DNSUpstream = config.Config.AlternativeDNSServer
		err := s.outbound.ExchangeFromRemote()
		if err != nil {
			log.Warn("Get dns response failed: ", err)
		}
		return
	}

	log.Debug("Finally use primary DNS")
}


