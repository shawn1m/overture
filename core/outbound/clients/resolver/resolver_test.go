package resolver

import (
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"
	"net"
	"os"
	"testing"
)

var questionDomain = "www.yahoo.com."
var udpUpstream = &common.DNSUpstream{
	Name:          "Test-UDP",
	Address:       "114.114.114.114",
	Protocol:      "udp",
	SOCKS5Address: "",
	Timeout:       6,
	EDNSClientSubnet: &common.EDNSClientSubnetType{
		Policy:     "disable",
		ExternalIP: "",
		NoCookie:   false,
	},
}

var tcpUpstream = &common.DNSUpstream{
	Name:          "Test-TCP",
	Address:       "1.1.1.1",
	Protocol:      "tcp",
	SOCKS5Address: "",
	Timeout:       6,
	EDNSClientSubnet: &common.EDNSClientSubnetType{
		Policy:     "disable",
		ExternalIP: "",
		NoCookie:   false,
	},
}

var tcpTlsUpstream = &common.DNSUpstream{
	Name:          "Test-TCP-TLS",
	Address:       "dns.google@8.8.8.8",
	Protocol:      "tcp-tls",
	SOCKS5Address: "",
	Timeout:       8,
	EDNSClientSubnet: &common.EDNSClientSubnetType{
		Policy:     "disable",
		ExternalIP: "",
		NoCookie:   false,
	},
}

var httpsUpstream = &common.DNSUpstream{
	Name:          "Test-HTTPS",
	Address:       "https://dns.google/dns-query",
	Protocol:      "https",
	SOCKS5Address: "",
	Timeout:       8,
	EDNSClientSubnet: &common.EDNSClientSubnetType{
		Policy:     "disable",
		ExternalIP: "",
		NoCookie:   false,
	},
}

func init() {
	os.Chdir("../..")
	//conf := config.NewConfig("config.test.json")
}

func TestDispatcher(t *testing.T) {
	testUDP(t)
	testTCP(t)
	testTCPTLS(t)
	testHTTPS(t)
}

func testUDP(t *testing.T) {
	q := getQueryMsg(questionDomain, dns.TypeA)
	resolver := NewResolver(udpUpstream)
	resp, err := resolver.Exchange(q)
	if err != nil {
		t.Errorf("Got error: %s", err)
	}
	if net.ParseIP(common.FindRecordByType(resp, dns.TypeA)).To4() == nil {
		t.Error(questionDomain + " should have A record")
	}
}

func testTCP(t *testing.T) {
	q := getQueryMsg(questionDomain, dns.TypeA)
	resolver := NewResolver(tcpUpstream)
	resp, _ := resolver.Exchange(q)
	if net.ParseIP(common.FindRecordByType(resp, dns.TypeA)).To4() == nil {
		t.Error(questionDomain + " should have A record")
	}
}

func testTCPTLS(t *testing.T) {
	q := getQueryMsg(questionDomain, dns.TypeA)
	resolver := NewResolver(tcpTlsUpstream)
	resp, _ := resolver.Exchange(q)
	if net.ParseIP(common.FindRecordByType(resp, dns.TypeA)).To4() == nil {
		t.Error(questionDomain + " should have A record")
	}
}

func testHTTPS(t *testing.T) {
	q := getQueryMsg(questionDomain, dns.TypeA)
	resolver := NewResolver(httpsUpstream)
	resp, _ := resolver.Exchange(q)
	if net.ParseIP(common.FindRecordByType(resp, dns.TypeA)).To4() == nil {
		t.Error(questionDomain + " should have A record")
	}
}

func getQueryMsg(z string, t uint16) *dns.Msg {
	q := new(dns.Msg)
	q.SetQuestion(z, t)
	return q
}
