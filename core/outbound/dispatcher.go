package outbound

import (
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/hosts"
)

type Dispatcher struct {
	QuestionMessage *dns.Msg

	PrimaryDNS     []*common.DNSUpstream
	AlternativeDNS []*common.DNSUpstream
	OnlyPrimaryDNS bool

	PrimaryClientBundle     *ClientBundle
	AlternativeClientBundle *ClientBundle
	ActiveClientBundle      *ClientBundle

	IPNetworkPrimaryList     []*net.IPNet
	IPNetworkAlternativeList []*net.IPNet
	DomainPrimaryList        []string
	DomainAlternativeList    []string
	RedirectIPv6Record       bool

	InboundIP string

	Hosts *hosts.Hosts
	Cache *cache.Cache
}

func (d *Dispatcher) Exchange() {

	d.PrimaryClientBundle = NewClientBundle(d.QuestionMessage, d.PrimaryDNS, d.InboundIP, d.Hosts, d.Cache)
	d.AlternativeClientBundle = NewClientBundle(d.QuestionMessage, d.AlternativeDNS, d.InboundIP, d.Hosts, d.Cache)

	for _, cb := range [2]*ClientBundle{d.PrimaryClientBundle, d.AlternativeClientBundle} {
		if ok := cb.ExchangeFromLocal(); ok {
			d.ActiveClientBundle = cb
			return
		}
	}

	if d.OnlyPrimaryDNS || d.ExchangeForPrimaryDomain() {
		d.ActiveClientBundle = d.PrimaryClientBundle
		d.ActiveClientBundle.ExchangeFromRemote(true, true)
		return
	}

	if ok := d.ExchangeForIPv6() || d.ExchangeForAlternativeDomain(); ok {
		d.ActiveClientBundle = d.AlternativeClientBundle
		d.ActiveClientBundle.ExchangeFromRemote(true, true)
		return
	}

	d.ChooseActiveClientBundle()
	if d.ActiveClientBundle == d.AlternativeClientBundle {
		d.ActiveClientBundle.ExchangeFromRemote(false, true)
	}
	d.ActiveClientBundle.CacheResult()
}

func (d *Dispatcher) ExchangeForIPv6() bool {

	if (d.PrimaryClientBundle.QuestionMessage.Question[0].Qtype == dns.TypeAAAA) && d.RedirectIPv6Record {
		d.ActiveClientBundle = d.AlternativeClientBundle
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) ExchangeForAlternativeDomain() bool {

	qn := d.PrimaryClientBundle.QuestionMessage.Question[0].Name[:len(d.PrimaryClientBundle.QuestionMessage.Question[0].Name)-1]

	for _, domain := range d.DomainAlternativeList {

		if qn == domain || strings.HasSuffix(qn, "."+domain) {
			log.Debug("Matched: Custom domain " + qn + " " + domain)
			d.ActiveClientBundle = d.AlternativeClientBundle
			log.Debug("Finally use alternative DNS")
			return true
		}
	}

	log.Debug("Domain match fail, try to use primary DNS")

	return false
}

func (d *Dispatcher) ExchangeForPrimaryDomain() bool {

	qn := d.PrimaryClientBundle.QuestionMessage.Question[0].Name[:len(d.PrimaryClientBundle.QuestionMessage.Question[0].Name)-1]

	for _, domain := range d.DomainPrimaryList {

		if qn == domain || strings.HasSuffix(qn, "."+domain) {
			log.Debug("Matched: Domain WhiteList " + qn + " " + domain)
			d.ActiveClientBundle = d.PrimaryClientBundle
			log.Debug("Finally use primary DNS")
			return true
		}
	}

	log.Debug("Domain white list match fail, try to use alternative DNS")

	return false
}

func (d *Dispatcher) ChooseActiveClientBundle() {

	d.ActiveClientBundle = d.PrimaryClientBundle
	d.PrimaryClientBundle.ExchangeFromRemote(false, true)

	if d.PrimaryClientBundle.ResponseMessage == nil || !common.HasAnswer(d.PrimaryClientBundle.ResponseMessage) {
		//log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		//d.ActiveClientBundle = d.AlternativeClientBundle
		return
	}

	for _, a := range d.PrimaryClientBundle.ResponseMessage.Answer {
		if a.Header().Rrtype == dns.TypeA {
			log.Debug("Try to match response ip address with IP network")
			if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), d.IPNetworkPrimaryList, true) {
				d.ActiveClientBundle = d.PrimaryClientBundle
				log.Debug("Primary")
				break
			}
			if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), d.IPNetworkAlternativeList, true) {
				d.ActiveClientBundle = d.AlternativeClientBundle
				log.Debug("Alternative")
				break
			}
		} else if a.Header().Rrtype == dns.TypeAAAA {
			log.Debug("Try to match response ip address with IP network")
			if common.IsIPMatchList(net.ParseIP(a.(*dns.AAAA).AAAA.String()), d.IPNetworkPrimaryList, true) {
				d.ActiveClientBundle = d.AlternativeClientBundle
				log.Debug("IPv6")
				break
			}
		} else {
			continue
		}
		return
	}

	log.Debug("Finally use primary DNS")
	d.ActiveClientBundle = d.PrimaryClientBundle
}
