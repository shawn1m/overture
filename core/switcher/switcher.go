package switcher

import (
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/common"
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/outbound"
	"github.com/miekg/dns"
)

type Switcher struct {
	ol                 *outbound.OutboundListType
	ipNetworkList      []*net.IPNet
	domainList         []string
	redirectIPv6Record bool
}

func NewSwitcher(outbound *outbound.OutboundListType) *Switcher {

	return &Switcher{
		ol:                 outbound,
		ipNetworkList:      config.Config.IPNetworkList,
		domainList:         config.Config.DomainList,
		redirectIPv6Record: config.Config.RedirectIPv6Record,
	}
}

func (s *Switcher) ChooseDNS() bool {

	qn := s.ol.QuestionMessage.Question[0].Name[:len(s.ol.QuestionMessage.Question[0].Name)-1]

	if common.IsQuestionInIPv6(s.ol.QuestionMessage) && s.redirectIPv6Record {
		s.ol.UpdateDNSUpstream(config.Config.AlternativeDNS)
		s.ol.ExchangeFromRemote(true)
		log.Debug("Finally use alternative DNS")
		return true
	}

	for _, d := range s.domainList {

		if qn == d || strings.HasSuffix(qn, "."+d) {
			log.Debug("Matched: Custom domain " + qn + " " + d)
			s.ol.UpdateDNSUpstream(config.Config.AlternativeDNS)
			s.ol.ExchangeFromRemote(true)
			log.Debug("Finally use alternative DNS")
			return true
		}
	}

	log.Debug("Domain match fail, try to use primary DNS")

	return false
}

func (s *Switcher) HandleResponseFromPrimaryDNS() {

	s.ol.ExchangeFromRemote(false)

	if s.ol.ResponseMessage == nil || len(s.ol.ResponseMessage.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		s.ol.UpdateDNSUpstream(config.Config.AlternativeDNS)
		s.ol.ExchangeFromRemote(true)
		return
	}

	for _, a := range s.ol.ResponseMessage.Answer {
		if a.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network")
		if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), s.ipNetworkList, true) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS")
		s.ol.UpdateDNSUpstream(config.Config.AlternativeDNS)
		s.ol.ExchangeFromRemote(true)
		return
	}

	go func() {
		s.ol.ExchangeFromRemote(true)
	}()

	log.Debug("Finally use primary DNS")
}
