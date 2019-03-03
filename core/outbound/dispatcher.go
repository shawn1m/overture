package outbound

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/outbound/clients"
)

type Dispatcher struct {
	QuestionMessage *dns.Msg
	ResponseMessage *dns.Msg

	PrimaryDNS     []*common.DNSUpstream
	AlternativeDNS []*common.DNSUpstream
	OnlyPrimaryDNS bool

	PrimaryClientBundle     *clients.RemoteClientBundle
	AlternativeClientBundle *clients.RemoteClientBundle
	ActiveClientBundle      *clients.RemoteClientBundle

	IPNetworkPrimaryList     []*net.IPNet
	IPNetworkAlternativeList []*net.IPNet
	DomainPrimaryList        []string
	DomainAlternativeList    []string
	RedirectIPv6Record       bool

	InboundIP  string
	MinimumTTL int

	Hosts *hosts.Hosts
	Cache *cache.Cache
}

func (d *Dispatcher) Exchange() *dns.Msg {

	d.PrimaryClientBundle = clients.NewClientBundle(d.QuestionMessage, d.PrimaryDNS, d.InboundIP, d.MinimumTTL, d.Cache, "Primary")
	d.AlternativeClientBundle = clients.NewClientBundle(d.QuestionMessage, d.AlternativeDNS, d.InboundIP, d.MinimumTTL, d.Cache, "Alternative")

	localClient := clients.NewLocalClient(d.QuestionMessage, d.Hosts)
	d.ResponseMessage = localClient.Exchange()
	if d.ResponseMessage != nil {
		return d.ResponseMessage
	}

	if d.OnlyPrimaryDNS || d.isSelectDomain(d.PrimaryClientBundle, d.DomainPrimaryList) {
		d.ActiveClientBundle = d.PrimaryClientBundle
		return d.ActiveClientBundle.Exchange(true, true)
	}

	if ok := d.isExchangeForIPv6() || d.isSelectDomain(d.AlternativeClientBundle, d.DomainAlternativeList); ok {
		d.ActiveClientBundle = d.AlternativeClientBundle
		return d.ActiveClientBundle.Exchange(true, true)
	}

	d.selectByIPNetwork()

	if d.ActiveClientBundle == d.AlternativeClientBundle {
		d.ResponseMessage = d.ActiveClientBundle.Exchange(false, true)
		d.ActiveClientBundle.CacheResult()
		return d.ResponseMessage
	}

	return d.PrimaryClientBundle.GetResponseMessage()
}

func (d *Dispatcher) isExchangeForIPv6() bool {

	if (d.PrimaryClientBundle.IsType(dns.TypeAAAA)) && d.RedirectIPv6Record {
		d.ActiveClientBundle = d.AlternativeClientBundle
		log.Debug("Finally use alternative DNS")
		return true
	}

	return false
}

func (d *Dispatcher) isSelectDomain(rcb *clients.RemoteClientBundle, dl []string) bool {

	qn := d.PrimaryClientBundle.GetFirstQuestionDomain()

	for _, domain := range dl {

		if common.IsDomainMatchRule(domain, qn) {
			log.WithFields(log.Fields{
				"DNS":      rcb.Name,
				"question": qn,
				"domain":   domain,
			}).Debug("Matched")
			d.ActiveClientBundle = rcb
			log.Debug("Finally use " + rcb.Name + " DNS")
			return true
		}
	}

	log.Debug("Domain " + rcb.Name + " match fail")

	return false
}

func (d *Dispatcher) selectByIPNetwork() {

	d.ActiveClientBundle = d.PrimaryClientBundle
	primaryResponse := d.PrimaryClientBundle.Exchange(false, true)

	if primaryResponse == nil || !common.HasAnswer(primaryResponse) {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		d.ActiveClientBundle = d.AlternativeClientBundle
		return
	}
	if d.PrimaryClientBundle.GetResponseMessage() == nil {
		log.Debug("d.PrimaryClientBundle.GetResponseMessage() is nil")
		d.ActiveClientBundle = d.AlternativeClientBundle
		return
	}
	for _, a := range d.PrimaryClientBundle.GetResponseMessage().Answer {
		log.Debug("Try to match response ip address with IP network")
		if a.Header().Rrtype == dns.TypeA {
			if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), d.IPNetworkPrimaryList, true, "primary") {
				d.ActiveClientBundle = d.PrimaryClientBundle
				log.Debug("Finally use primary DNS")
				return
			}
			if common.IsIPMatchList(net.ParseIP(a.(*dns.A).A.String()), d.IPNetworkAlternativeList, true, "alternative") {
				d.ActiveClientBundle = d.AlternativeClientBundle
				log.Debug("Finally use alternative DNS")
				return
			}
		} else if a.Header().Rrtype == dns.TypeAAAA {
			if common.IsIPMatchList(net.ParseIP(a.(*dns.AAAA).AAAA.String()), d.IPNetworkPrimaryList, true, "primary") {
				d.ActiveClientBundle = d.PrimaryClientBundle
				log.Debug("Finally use primary DNS")
				return
			}
			if common.IsIPMatchList(net.ParseIP(a.(*dns.AAAA).AAAA.String()), d.IPNetworkAlternativeList, true, "alternative") {
				d.ActiveClientBundle = d.AlternativeClientBundle
				log.Debug("Finally use alternative DNS")
				return
			}
		}
		log.Debug("IP network match failed, finally use alternative DNS")
		d.ActiveClientBundle = d.AlternativeClientBundle
	}
}
