package outbound

import (
	"net"

	"github.com/holyshawn/overture/core/common"
	"github.com/holyshawn/overture/core/config"
	"github.com/miekg/dns"
)

func (o *Outbound) GetEDNSClientSubnetIP() string {

	switch o.DNSUpstream.EDNSClientSubnetPolicy {
	case "auto":
		if !common.IsIPMatchList(net.ParseIP(o.inboundIP), config.Config.ReservedIPNetworkList, false) {
			return o.inboundIP
		} else {
			return o.DNSUpstream.EDNSClientSubnetExternalIP
		}
	case "disable":
	}
	return ""
}

func setEDNSClientSubnet(m *dns.Msg, ip string) {

	if ip == "" {
		return
	}

	o := m.IsEdns0()
	if o == nil {
		o = new(dns.OPT)
		m.Extra = append(m.Extra, o)
	}
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT

	es := IsEDNSClientSubnet(o)
	if es == nil {
		es = new(dns.EDNS0_SUBNET)
		o.Option = append(o.Option, es)
	}
	es.Code = dns.EDNS0SUBNET
	es.Address = net.ParseIP(ip)
	if es.Address.To4() != nil {
		es.Family = 1         // 1 for IPv4 source address, 2 for IPv6
		es.SourceNetmask = 32 // 32 for IPV4, 128 for IPv6
	} else {
		es.Family = 2          // 1 for IPv4 source address, 2 for IPv6
		es.SourceNetmask = 128 // 32 for IPV4, 128 for IPv6
	}
	es.SourceScope = 0
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
