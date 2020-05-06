package common

import (
	"net"
	"testing"
)

func contains(ipNets []*net.IPNet, ip net.IP) bool {
	for _, ipNet := range ipNets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func TestIPSet(t *testing.T) {
	var ipNetList []*net.IPNet
	for _, s := range []string{
		"::ffff:192.168.1.0/120",
		"192.168.2.0/24",
		"192.168.4.0/24",
		"::fffe:0:0/95", // a range covers all IPv4-mapped Address
		"fe80::1234:5678:9000/120",
	} {
		_, ipNet, _ := net.ParseCIDR(s)
		ipNetList = append(ipNetList, ipNet)
	}
	ipSet := NewIPSet(ipNetList)

	for _, s := range []string{
		"192.168.3.0",
		"192.168.2.1",
		"::ffff:192.168.2.1",
		"192.168.1.1",
		"fe80::1234:5678:8fff",
		"fe80::1234:5678:9000",
		"fe80::1234:5678:90ff",
		"fe80::1234:5678:9100",
		"invalid ip",
	} {
		ip := net.ParseIP(s)
		expect := contains(ipNetList, ip)
		result := ipSet.Contains(ip, true, s)
		if expect != result {
			t.Errorf("expect %v, but got %v: '%v'", expect, result, s)
		}
	}
}
