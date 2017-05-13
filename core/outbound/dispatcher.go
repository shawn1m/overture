package outbound

import (
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"
)

type Dispatcher struct {
	PrimaryDNS     []*DNSUpstream
	AlternativeDNS []*DNSUpstream
	OnlyPrimaryDNS bool

	ClientBundle       *ClientBundle
	IPNetworkList      []*net.IPNet
	DomainList         []string
	RedirectIPv6Record bool
}

func (d *Dispatcher) Exchange() {

	if ok := d.ClientBundle.ExchangeFromLocal(); ok {
		return
	}

	if d.OnlyPrimaryDNS {
		d.ClientBundle.ExchangeFromRemote(true, true)
		d.ClientBundle.CacheResults()
		return
	}

	if ok := d.ExchangeForIPv6() || d.ExchangeForDomain(); ok {
		d.ClientBundle.CacheResults()
		return
	}

	d.ClientBundle.ExchangeFromRemote(false, true)
	d.ExchangeForPrimaryDNSResponse()
	d.ClientBundle.CacheResults()
}

func (d *Dispatcher) ExchangeForIPv6() bool {

	if (d.ClientBundle.QuestionMessage.Question[0].Qtype == dns.TypeAAAA) && d.RedirectIPv6Record {
		d.ClientBundle.UpdateFromDNSUpstream(d.AlternativeDNS)
		d.ClientBundle.ExchangeFromRemote(true, true)
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) ExchangeForDomain() bool {

	qn := d.ClientBundle.QuestionMessage.Question[0].Name[:len(d.ClientBundle.QuestionMessage.Question[0].Name)-1]

	for _, domain := range d.DomainList {

		if qn == domain || strings.HasSuffix(qn, "."+domain) {
			log.Debug("Matched: Custom domain " + qn + " " + domain)
			d.ClientBundle.UpdateFromDNSUpstream(d.AlternativeDNS)
			d.ClientBundle.ExchangeFromRemote(true, true)
			log.Debug("Finally use alternative DNS")
			return true
		}
	}

	log.Debug("Domain match fail, try to use primary DNS")

	return false
}

func (d *Dispatcher) ExchangeForPrimaryDNSResponse() {

	if d.ClientBundle.ResponseMessage == nil || len(d.ClientBundle.ResponseMessage.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		d.ClientBundle.UpdateFromDNSUpstream(d.AlternativeDNS)
		d.ClientBundle.ExchangeFromRemote(true, true)
		return
	}

	for _, a := range d.ClientBundle.ResponseMessage.Answer {
		if a.Header().Rrtype == dns.TypeA {
			log.Debug("Try to match response ip address with IP network")
			if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), d.IPNetworkList, true) {
				break
			}
		} else if a.Header().Rrtype == dns.TypeAAAA {
			log.Debug("Try to match response ip address with IP network")
			if common.IsIPMatchList(net.ParseIP(a.(*dns.AAAA).AAAA.String()), d.IPNetworkList, true) {
				break
			}
		} else {
			continue
		}

		log.Debug("IP network match fail, finally use alternative DNS")
		d.ClientBundle.UpdateFromDNSUpstream(d.AlternativeDNS)
		d.ClientBundle.ExchangeFromRemote(true, true)
		return
	}

	log.Debug("Finally use primary DNS")
}
