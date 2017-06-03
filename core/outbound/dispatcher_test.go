package outbound

import (
	"testing"
	"os"

	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/config"
	"github.com/shawn1m/overture/core/common"
)

func TestDispatcher_Exchange(t *testing.T) {

	os.Chdir("../..")
	c := config.NewConfig("config.sample.json")
	d := &Dispatcher{
		PrimaryDNS:         c.PrimaryDNS,
		AlternativeDNS:     c.AlternativeDNS,
		OnlyPrimaryDNS:     c.OnlyPrimaryDNS,
		IPNetworkList:      c.IPNetworkList,
		DomainList:         c.DomainList,
		RedirectIPv6Record: c.RedirectIPv6Record,
	}

	inboundIP := ""
	q := new(dns.Msg)
	q.SetQuestion("www.baidu.com.", dns.TypeA)
	d.PrimaryClientBundle = NewClientBundle(q, d.PrimaryDNS, inboundIP, c.Hosts, c.Cache)
	d.AlternativeClientBundle = NewClientBundle(q, d.AlternativeDNS, inboundIP, c.Hosts, c.Cache)
	d.Exchange()
	println(common.FindIPByType(d.ActiveClientBundle.ResponseMessage, dns.TypeA))
	if common.FindIPByType(d.ActiveClientBundle.ResponseMessage, dns.TypeA) == "" {
		t.Error("baidu.com should have an A record")
	}

	q.SetQuestion("www.twitter.com.", dns.TypeAAAA)
	d.PrimaryClientBundle = NewClientBundle(q, d.PrimaryDNS, inboundIP, c.Hosts, c.Cache)
	d.AlternativeClientBundle = NewClientBundle(q, d.AlternativeDNS, inboundIP, c.Hosts, c.Cache)
	d.Exchange()
	println(common.FindIPByType(d.ActiveClientBundle.ResponseMessage, dns.TypeAAAA))
	if common.FindIPByType(d.ActiveClientBundle.ResponseMessage, dns.TypeAAAA) != "" {
		t.Error("twitter.com should't have AAAA record")
	}
}
