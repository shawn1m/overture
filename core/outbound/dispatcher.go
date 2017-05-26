package outbound

import (
	"net"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"
)

type Dispatcher struct {
	PrimaryDNS     []*DNSUpstream
	AlternativeDNS []*DNSUpstream
	OnlyPrimaryDNS bool

	PrimaryClientBundle     *ClientBundle
	AlternativeClientBundle *ClientBundle
	ActiveClientBundle      *ClientBundle

	IPNetworkList      []*net.IPNet
	DomainList         []string
	RedirectIPv6Record bool
}

func (d *Dispatcher) Exchange() {

	for _, cb := range [2]*ClientBundle{d.PrimaryClientBundle, d.AlternativeClientBundle} {
		if ok := cb.ExchangeFromLocal(); ok {
			d.ActiveClientBundle = cb
			return
		}
	}

	if d.OnlyPrimaryDNS {
		d.PrimaryClientBundle.ExchangeFromRemote(true, true)
		d.ActiveClientBundle = d.PrimaryClientBundle
		return
	}

	var awg sync.WaitGroup
	awg.Add(1)
	go func() {
		d.AlternativeClientBundle.ExchangeFromRemote(false, true)
		awg.Done()
	}()

	if ok := d.ExchangeForIPv6() || d.ExchangeForDomain(); ok {
		awg.Wait()
		d.ActiveClientBundle.CacheResult()
		return
	}

	var pwg sync.WaitGroup
	pwg.Add(1)
	go func() {
		d.PrimaryClientBundle.ExchangeFromRemote(false, true)
		pwg.Done()
	}()

	pwg.Wait()
	d.ExchangeForPrimaryDNSResponse()
	if d.ActiveClientBundle == d.AlternativeClientBundle {
		awg.Wait()
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

func (d *Dispatcher) ExchangeForDomain() bool {

	qn := d.PrimaryClientBundle.QuestionMessage.Question[0].Name[:len(d.PrimaryClientBundle.QuestionMessage.Question[0].Name)-1]

	for _, domain := range d.DomainList {

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

func (d *Dispatcher) ExchangeForPrimaryDNSResponse() {

	if d.PrimaryClientBundle.ResponseMessage == nil || len(d.PrimaryClientBundle.ResponseMessage.Answer) == 0 {
		log.Debug("Primary DNS answer is empty, finally use alternative DNS")
		d.ActiveClientBundle = d.AlternativeClientBundle
		return
	}

	for _, a := range d.PrimaryClientBundle.ResponseMessage.Answer {
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
		d.ActiveClientBundle = d.AlternativeClientBundle
		return
	}

	log.Debug("Finally use primary DNS")
	d.ActiveClientBundle = d.PrimaryClientBundle
}
