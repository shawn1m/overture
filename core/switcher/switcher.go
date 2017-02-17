// Copyright (c) 2016 holyshawn. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

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
	ob                 *outbound.OutboundBundle
	ipNetworkList      []*net.IPNet
	domainList         []string
	redirectIPv6Record bool
}

func NewSwitcher(outbound *outbound.OutboundBundle) *Switcher {

	return &Switcher{
		ob:                 outbound,
		ipNetworkList:      config.Config.IPNetworkList,
		domainList:         config.Config.DomainList,
		redirectIPv6Record: config.Config.RedirectIPv6Record,
	}
}

func (s *Switcher) ExchangeForIPv6() bool {

	if (s.ob.QuestionMessage.Question[0].Qtype == dns.TypeAAAA) && s.redirectIPv6Record {
		s.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
		s.ob.ExchangeFromRemote(true, true)
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (s *Switcher) ExchangeForDomain() bool {

	qn := s.ob.QuestionMessage.Question[0].Name[:len(s.ob.QuestionMessage.Question[0].Name)-1]

	for _, d := range s.domainList {

		if qn == d || strings.HasSuffix(qn, "."+d) {
			log.Debug("Matched: Custom domain " + qn + " " + d)
			s.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
			s.ob.ExchangeFromRemote(true, true)
			log.Debug("Finally use alternative DNS")
			return true
		}
	}

	log.Debug("Domain match fail, try to use primary DNS")

	return false
}

func (s *Switcher) ExchangeForPrimaryDNSResponse() {

	if s.ob.ResponseMessage == nil || len(s.ob.ResponseMessage.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		s.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
		s.ob.ExchangeFromRemote(true, true)
		return
	}

	for _, a := range s.ob.ResponseMessage.Answer {
		if a.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network")
		if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), s.ipNetworkList, true) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS")
		s.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
		s.ob.ExchangeFromRemote(true, true)
		return
	}

	go func() {
		s.ob.ExchangeFromRemote(true, false)
	}()

	log.Debug("Finally use primary DNS")
}
