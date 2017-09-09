package outbound

import (
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/config"
)

var c *config.Config
var d Dispatcher

func init() {
	os.Chdir("../..")
	c = config.NewConfig("config.test.json")
	d = Dispatcher{
		PrimaryDNS:         c.PrimaryDNS,
		AlternativeDNS:     c.AlternativeDNS,
		OnlyPrimaryDNS:     c.OnlyPrimaryDNS,
		IPNetworkList:      c.IPNetworkList,
		DomainList:         c.DomainList,
		RedirectIPv6Record: c.RedirectIPv6Record,
		InboundIP:          "",
		Hosts:              c.Hosts,
		Cache:              c.Cache,
	}
}

func TestDispatcher(t *testing.T) {

	testHosts(t)
	testIPResponse(t)

	testAAAA(t)
	testForeign(t)

	d.DomainList = nil
	testDomestic(t)
	testForeign(t)

	testCache(t)
}

func testDomestic(t *testing.T) {

	exchange("www.baidu.com.", dns.TypeA)
	if common.FindRecordByType(d.ActiveClientBundle.ResponseMessage, dns.TypeA) == "" {
		t.Error("baidu.com should have an A record")
	}
}

func testForeign(t *testing.T) {

	exchange("www.twitter.com.", dns.TypeA)
	if common.FindRecordByType(d.ActiveClientBundle.ResponseMessage, dns.TypeCNAME) != "twitter.com." {
		t.Error("twitter.com should have an twitter.com CNAME record")
	}
}

func testAAAA(t *testing.T) {

	exchange("www.twitter.com.", dns.TypeAAAA)
	if common.FindRecordByType(d.ActiveClientBundle.ResponseMessage, dns.TypeAAAA) != "" {
		t.Error("twitter.com should not have AAAA record")
	}
}

func testHosts(t *testing.T) {

	exchange("localhost.", dns.TypeA)
	if common.FindRecordByType(d.ActiveClientBundle.ResponseMessage, dns.TypeA) != "127.0.0.1" {
		t.Error("localhost should be 127.0.0.1")
	}
}

func testIPResponse(t *testing.T) {

	exchange("127.0.0.1.", dns.TypeA)
	if common.FindRecordByType(d.ActiveClientBundle.ResponseMessage, dns.TypeA) != "127.0.0.1" {
		t.Error("127.0.0.1 should be 127.0.0.1")
	}

	exchange("fe80::7f:4f42:3f4d:f4c8.", dns.TypeAAAA)
	if common.FindRecordByType(d.ActiveClientBundle.ResponseMessage, dns.TypeAAAA) != "fe80::7f:4f42:3f4d:f4c8" {
		t.Error("fe80::7f:4f42:3f4d:f4c8 should be fe80::7f:4f42:3f4d:f4c8")
	}
}

func testCache(t *testing.T) {

	exchange("www.cnn.com.", dns.TypeA)
	now := time.Now()
	d.Exchange()
	if time.Since(now) > 10*time.Millisecond {
		t.Error("Cache response slower than 10ms")
	}
}

func exchange(z string, t uint16) {

	q := new(dns.Msg)
	q.SetQuestion(z, t)
	d.QuestionMessage = q
	d.Exchange()
}
