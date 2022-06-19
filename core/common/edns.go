package common

import (
	"net"

	"github.com/miekg/dns"
)

type EDNSClientSubnetType struct {
	Policy     string `yaml:"policy" json:"policy"`
	ExternalIP string `yaml:"externalIP" json:"externalIP"`
	NoCookie   bool   `yaml:"noCookie"json:"noCookie"`
}

func SetEDNSClientSubnet(m *dns.Msg, ip string, isNoCookie bool) {
	if ip == "" {
		return
	}

	o := m.IsEdns0()
	if o == nil {
		o = new(dns.OPT)
		o.SetUDPSize(4096)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
		m.Extra = append(m.Extra, o)
	}

	es := IsEDNSClientSubnet(o)
	if es == nil || es.Address.IsUnspecified() {
		nes := new(dns.EDNS0_SUBNET)
		nes.Code = dns.EDNS0SUBNET
		nes.Address = net.ParseIP(ip)
		if nes.Address.To4() != nil {
			nes.Family = 1         // 1 for IPv4 source address, 2 for IPv6
			nes.SourceNetmask = 24 // 24 for IPV4, 56 for IPv6
		} else {
			nes.Family = 2         // 1 for IPv4 source address, 2 for IPv6
			nes.SourceNetmask = 56 // 24 for IPV4, 56 for IPv6
		}
		nes.SourceScope = 0
		if es != nil && es.Address.IsUnspecified() {
			var edns0 []dns.EDNS0
			for _, s := range o.Option {
				switch e := s.(type) {
				case *dns.EDNS0_SUBNET:
				default:
					edns0 = append(edns0, e)
				}
			}
			o.Option = edns0
		}
		o.Option = append(o.Option, nes)
		if isNoCookie {
			deleteCookie(o)
		}
	}
}

func deleteCookie(o *dns.OPT) {
	for i, e0 := range o.Option {
		switch e0.(type) {
		case *dns.EDNS0_COOKIE:
			o.Option = append(o.Option[:i], o.Option[i+1:]...)
		}
	}
}

func IsEDNSClientSubnet(o *dns.OPT) *dns.EDNS0_SUBNET {
	for _, s := range o.Option {
		switch e := s.(type) {
		case *dns.EDNS0_SUBNET:
			return e
		}
	}
	return nil
}

func GetEDNSClientSubnetIP(m *dns.Msg) string {
	o := m.IsEdns0()
	if o != nil {
		for _, s := range o.Option {
			switch e := s.(type) {
			case *dns.EDNS0_SUBNET:
				return e.Address.String()
			}
		}
	}
	return ""
}
