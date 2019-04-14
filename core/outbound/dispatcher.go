package outbound

import (
	"github.com/shawn1m/overture/core/matcher"
	"net"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/outbound/clients"
)

type Dispatcher struct {
	PrimaryDNS     []*common.DNSUpstream
	AlternativeDNS []*common.DNSUpstream
	OnlyPrimaryDNS bool

	WhenPrimaryDNSAnswerNoneUse string
	IPNetworkPrimaryList        []*net.IPNet
	IPNetworkAlternativeList    []*net.IPNet
	DomainPrimaryList           matcher.Matcher
	DomainAlternativeList       matcher.Matcher
	RedirectIPv6Record          bool

	MinimumTTL   int
	DomainTTLMap map[string]uint32

	Hosts *hosts.Hosts
	Cache *cache.Cache
}

func (d *Dispatcher) Exchange(query *dns.Msg, inboundIP string) *dns.Msg {

	PrimaryClientBundle := clients.NewClientBundle(query, d.PrimaryDNS, inboundIP, d.MinimumTTL, d.Cache, "Primary", d.DomainTTLMap)
	AlternativeClientBundle := clients.NewClientBundle(query, d.AlternativeDNS, inboundIP, d.MinimumTTL, d.Cache, "Alternative", d.DomainTTLMap)
	var ActiveClientBundle *clients.RemoteClientBundle

	localClient := clients.NewLocalClient(query, d.Hosts, d.MinimumTTL, d.DomainTTLMap)
	resp := localClient.Exchange()
	if resp != nil {
		return resp
	}

	for _, cb := range []*clients.RemoteClientBundle{PrimaryClientBundle, AlternativeClientBundle} {
		resp := cb.ExchangeFromCache()
		if resp != nil {
			return resp
		}
	}

	if resp != nil {
		return resp
	}

	if d.OnlyPrimaryDNS || d.isSelectDomain(PrimaryClientBundle, d.DomainPrimaryList) {
		ActiveClientBundle = PrimaryClientBundle
		return ActiveClientBundle.Exchange(true, true)
	}

	if ok := d.isExchangeForIPv6(query) || d.isSelectDomain(AlternativeClientBundle, d.DomainAlternativeList); ok {
		ActiveClientBundle = AlternativeClientBundle
		return ActiveClientBundle.Exchange(true, true)
	}

	ActiveClientBundle = d.selectByIPNetwork(PrimaryClientBundle, AlternativeClientBundle)

	if ActiveClientBundle == AlternativeClientBundle {
		resp = ActiveClientBundle.Exchange(true, true)
		return resp
	} else {
		// Only try to Cache result before return
		ActiveClientBundle.CacheResultIfNeeded()
		return ActiveClientBundle.GetResponseMessage()
	}
}

func (d *Dispatcher) isExchangeForIPv6(query *dns.Msg) bool {

	if query.Question[0].Qtype == dns.TypeAAAA && d.RedirectIPv6Record {
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) isSelectDomain(rcb *clients.RemoteClientBundle, dt matcher.Matcher) bool {

	qn := rcb.GetFirstQuestionDomain()

	if dt.Has(qn) {
		log.WithFields(log.Fields{
			"DNS":      rcb.Name,
			"question": qn,
			"domain":   qn,
		}).Debug("Matched")
		log.Debug("Finally use " + rcb.Name + " DNS")
		return true
	}

	log.Debug("Domain " + rcb.Name + " match fail")

	return false
}

func (d *Dispatcher) selectByIPNetwork(PrimaryClientBundle, AlternativeClientBundle *clients.RemoteClientBundle) *clients.RemoteClientBundle {

	primaryResponse := PrimaryClientBundle.Exchange(false, true)

	if primaryResponse == nil {
		log.Debug("Primary DNS return nil, finally use alternative DNS")
		return AlternativeClientBundle
	}

	if primaryResponse.Answer == nil {
		if d.WhenPrimaryDNSAnswerNoneUse == "AlternativeDNS" {
			log.Debug("Primary DNS response has no answer section but exist, finally use AlternativeDNS")
			return AlternativeClientBundle
		} else {
			log.Debug("Primary DNS response has no answer section but exist, finally use PrimaryDNS")
			return PrimaryClientBundle
		}
	}

	for _, a := range PrimaryClientBundle.GetResponseMessage().Answer {
		log.Debug("Try to match response ip address with IP network")
		var ip net.IP
		if a.Header().Rrtype == dns.TypeA {
			ip = net.ParseIP(a.(*dns.A).A.String())
		} else if a.Header().Rrtype == dns.TypeAAAA {
			ip = net.ParseIP(a.(*dns.AAAA).AAAA.String())
		} else {
			continue
		}
		if common.IsIPMatchList(ip, d.IPNetworkPrimaryList, true, "primary") {
			log.Debug("Finally use primary DNS")
			return PrimaryClientBundle
		}
		if common.IsIPMatchList(ip, d.IPNetworkAlternativeList, true, "alternative") {
			log.Debug("Finally use alternative DNS")
			return AlternativeClientBundle
		}
	}
	log.Debug("IP network match failed, finally use alternative DNS")
	return AlternativeClientBundle
}
