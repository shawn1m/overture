package outbound

import (
	"github.com/shawn1m/overture/core/outbound/clients/resolver"
	"net"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/matcher"
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
	AlternativeDNSConcurrent    bool
	PoolIdleTimeout          int

	MinimumTTL   int
	DomainTTLMap map[string]uint32

	Hosts *hosts.Hosts
	Cache *cache.Cache

	primaryResolvers     []resolver.Resolver
	alternativeResolvers []resolver.Resolver
}

func createResolver(ul []*common.DNSUpstream) (resolvers []resolver.Resolver) {
	resolvers = make([]resolver.Resolver, len(ul))
	for i, u := range ul {
		resolvers[i] = resolver.NewResolver(u)
	}
	return resolvers
}

func (d *Dispatcher) Init() {
	resolver.IdleTimeout = time.Duration(d.PoolIdleTimeout) * time.Second
	log.Debugf("Set pool's IdleTimeout to %d", d.PoolIdleTimeout)
	d.primaryResolvers = createResolver(d.PrimaryDNS)
	d.alternativeResolvers = createResolver(d.AlternativeDNS)
}

func (d *Dispatcher) Exchange(query *dns.Msg, inboundIP string) *dns.Msg {
	PrimaryClientBundle := clients.NewClientBundle(query, d.PrimaryDNS, d.primaryResolvers, inboundIP, d.MinimumTTL, d.Cache, "Primary", d.DomainTTLMap)
	AlternativeClientBundle := clients.NewClientBundle(query, d.AlternativeDNS, d.alternativeResolvers, inboundIP, d.MinimumTTL, d.Cache, "Alternative", d.DomainTTLMap)

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

	if d.OnlyPrimaryDNS || d.isSelectDomain(PrimaryClientBundle, d.DomainPrimaryList) {
		ActiveClientBundle = PrimaryClientBundle
		return ActiveClientBundle.Exchange(true, true)
	}

	if ok := d.isExchangeForIPv6(query) || d.isSelectDomain(AlternativeClientBundle, d.DomainAlternativeList); ok {
		ActiveClientBundle = AlternativeClientBundle
		return ActiveClientBundle.Exchange(true, true)
	}

	ActiveClientBundle = d.selectByIPNetwork(PrimaryClientBundle, AlternativeClientBundle)

	// Only try to Cache result before return
	ActiveClientBundle.CacheResultIfNeeded()
	return ActiveClientBundle.GetResponseMessage()
}

func (d *Dispatcher) isExchangeForIPv6(query *dns.Msg) bool {
	if query.Question[0].Qtype == dns.TypeAAAA && d.RedirectIPv6Record {
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) isSelectDomain(rcb *clients.RemoteClientBundle, dt matcher.Matcher) bool {
	if dt != nil {
		qn := rcb.GetFirstQuestionDomain()

		if dt.Has(qn) {
			log.WithFields(log.Fields{
				"DNS":      rcb.Name,
				"question": qn,
				"domain":   qn,
			}).Debug("Matched")
			log.Debugf("Finally use %s DNS", rcb.Name)
			return true
		}

		log.Debugf("Domain %s match fail", rcb.Name)
	} else {
		log.Debug("Domain matcher is nil, not checking")
	}

	return false
}

func (d *Dispatcher) selectByIPNetwork(PrimaryClientBundle, AlternativeClientBundle *clients.RemoteClientBundle) *clients.RemoteClientBundle {
	primaryOut := make(chan *dns.Msg)
	alternateOut := make(chan *dns.Msg)
	go func() {
		primaryOut <- PrimaryClientBundle.Exchange(false, true)
	}()
	alternateFunc := func() {
		alternateOut <- AlternativeClientBundle.Exchange(false, true)
	}
	waitAlternateResp := func() {
		if !d.AlternativeDNSConcurrent {
			go alternateFunc()
		}
		<-alternateOut
	}
	if d.AlternativeDNSConcurrent {
		go alternateFunc()
	}
	primaryResponse := <-primaryOut

	if primaryResponse != nil {
		if primaryResponse.Answer == nil {
			if d.WhenPrimaryDNSAnswerNoneUse != "AlternativeDNS" {
				log.Debug("Primary DNS response has no answer section but exist, finally use PrimaryDNS")
				return PrimaryClientBundle
			} else {
				log.Debug("Primary DNS response has no answer section but exist, finally use AlternativeDNS")
				waitAlternateResp()
				return AlternativeClientBundle
			}
		}
	} else {
		log.Debug("Primary DNS return nil, finally use alternative DNS")
		waitAlternateResp()
		return AlternativeClientBundle
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
			waitAlternateResp()
			return AlternativeClientBundle
		}
	}
	log.Debug("IP network match failed, finally use alternative DNS")
	waitAlternateResp()
	return AlternativeClientBundle
}
