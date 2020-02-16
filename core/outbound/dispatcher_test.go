package outbound

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"

	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/config"
)

var dispatcher Dispatcher
var questionDomain = "www.yahoo.com."

func init() {
	os.Chdir("../..")
	conf := config.NewConfig("config.test.json")
	dispatcher = Dispatcher{
		PrimaryDNS:                  conf.PrimaryDNS,
		AlternativeDNS:              conf.AlternativeDNS,
		OnlyPrimaryDNS:              conf.OnlyPrimaryDNS,
		WhenPrimaryDNSAnswerNoneUse: conf.WhenPrimaryDNSAnswerNoneUse,
		IPNetworkPrimaryList:        conf.IPNetworkPrimaryList,
		IPNetworkAlternativeList:    conf.IPNetworkAlternativeList,
		DomainPrimaryList:           conf.DomainPrimaryList,
		DomainAlternativeList:       conf.DomainAlternativeList,

		RedirectIPv6Record:       conf.IPv6UseAlternativeDNS,
		AlternativeDNSConcurrent: conf.AlternativeDNSConcurrent,
		PoolIdleTimeout:          conf.PoolIdleTimeout,
		MinimumTTL:               conf.MinimumTTL,
		DomainTTLMap:             conf.DomainTTLMap,

		Hosts: conf.Hosts,
		Cache: conf.Cache,
	}
	dispatcher.Init()
}

func TestDispatcher(t *testing.T) {

	testA(t)
	testAAAA(t)
	testHosts(t)
	testIPResponse(t)
	testCache(t)
}

func testA(t *testing.T) {

	resp := exchange(questionDomain, dns.TypeA)
	if net.ParseIP(common.FindRecordByType(resp, dns.TypeA)).To4() == nil {
		t.Error(questionDomain + " should have A record")
	}
}

func testAAAA(t *testing.T) {

	resp := exchange(questionDomain, dns.TypeAAAA)
	if net.ParseIP(common.FindRecordByType(resp, dns.TypeAAAA)).To16() == nil {
		t.Error(questionDomain + " should have AAAA record")
	}
}

func testHosts(t *testing.T) {

	resp := exchange("localhost.", dns.TypeA)
	if common.FindRecordByType(resp, dns.TypeA) != "127.0.0.1" {
		t.Error("localhost should be 127.0.0.1")
	}
}

func testIPResponse(t *testing.T) {

	resp := exchange("127.0.0.1.", dns.TypeA)
	if common.FindRecordByType(resp, dns.TypeA) != "127.0.0.1" {
		t.Error("127.0.0.1 should be 127.0.0.1")
	}

	resp = exchange("fe80::7f:4f42:3f4d:f4c8.", dns.TypeAAAA)
	if common.FindRecordByType(resp, dns.TypeAAAA) != "fe80::7f:4f42:3f4d:f4c8" {
		t.Error("fe80::7f:4f42:3f4d:f4c8 should be fe80::7f:4f42:3f4d:f4c8")
	}
}

func testCache(t *testing.T) {

	exchange(questionDomain, dns.TypeA)
	now := time.Now()
	exchange(questionDomain, dns.TypeA)
	if time.Since(now) > 10*time.Millisecond {
		t.Error(time.Since(now).String() + " " + "Cache response slower than 10ms")
	}
}

func exchange(z string, t uint16) *dns.Msg {

	q := new(dns.Msg)
	q.SetQuestion(z, t)
	return dispatcher.Exchange(q, "")
}
