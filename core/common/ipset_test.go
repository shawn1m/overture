package common

import (
	log "github.com/sirupsen/logrus"
	"net"
	"testing"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestNewIPSet(t *testing.T) {
	var ipNetList []*net.IPNet
	for _, s := range []string{
		"::ffff:192.168.1.0/120",
		"192.168.2.0/24",
		"192.168.4.0/24",
	} {
		_, ipNet, _ := net.ParseCIDR(s)
		ipNetList = append(ipNetList, ipNet)
	}
	ipSet := NewIPSet(ipNetList)
	if ipSet.Contains(net.ParseIP("192.168.3.0"), true, "test1") {
		t.Error()
	}
	if !ipSet.Contains(net.ParseIP("192.168.2.1"), true, "test2") {
		t.Error()
	}
	if !ipSet.Contains(net.ParseIP("::ffff:192.168.2.1"), true, "test3") {
		t.Error()
	}
	if !ipSet.Contains(net.ParseIP("192.168.1.1"), true, "test4") {
		t.Error()
	}
}
