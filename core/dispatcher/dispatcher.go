// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package dispatcher

import (
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/config"
	"github.com/shawn1m/overture/core/outbound"
)

type Dispatcher struct {
	ob                 *outbound.OutboundBundle
	ipNetworkList      []*net.IPNet
	domainList         []string
	redirectIPv6Record bool
}

func New(outbound *outbound.OutboundBundle) *Dispatcher {

	return &Dispatcher{
		ob:                 outbound,
		ipNetworkList:      config.Config.IPNetworkList,
		domainList:         config.Config.DomainList,
		redirectIPv6Record: config.Config.RedirectIPv6Record,
	}
}

func (d *Dispatcher) ExchangeForIPv6() bool {

	if (d.ob.QuestionMessage.Question[0].Qtype == dns.TypeAAAA) && d.redirectIPv6Record {
		d.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
		d.ob.ExchangeFromRemote(true, true)
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) ExchangeForDomain() bool {

	qn := d.ob.QuestionMessage.Question[0].Name[:len(d.ob.QuestionMessage.Question[0].Name)-1]

	for _, domain := range d.domainList {

		if qn == domain || strings.HasSuffix(qn, "."+ domain) {
			log.Debug("Matched: Custom domain " + qn + " " + domain)
			d.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
			d.ob.ExchangeFromRemote(true, true)
			log.Debug("Finally use alternative DNS")
			return true
		}
	}

	log.Debug("Domain match fail, try to use primary DNS")

	return false
}

func (d *Dispatcher) ExchangeForPrimaryDNSResponse() {

	if d.ob.ResponseMessage == nil || len(d.ob.ResponseMessage.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		d.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
		d.ob.ExchangeFromRemote(true, true)
		return
	}

	for _, a := range d.ob.ResponseMessage.Answer {
		if a.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network")
		if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), d.ipNetworkList, true) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS")
		d.ob.UpdateFromDNSUpstream(config.Config.AlternativeDNS)
		d.ob.ExchangeFromRemote(true, true)
		return
	}

	go func() {
		d.ob.ExchangeFromRemote(true, false)
	}()

	log.Debug("Finally use primary DNS")
}
