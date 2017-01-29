package outbound

import (
	"errors"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/common"
	"github.com/holyshawn/overture/core/config"
	"github.com/miekg/dns"
)

type Outbound struct {
	ResponseMessage        *dns.Msg
	QuestionMessage        *dns.Msg
	inboundIP              string
	DNSUpstream            *config.DNSUpstream
	EDNSClientSubnetPolicy string
	EDNSClientSubnetIP     string
	externalIP             string
	MinimumTTL             int
	IPUsed                 string
}

func NewOutbound(q *dns.Msg, inboundIP string) *Outbound {

	o := &Outbound{
		ResponseMessage:        new(dns.Msg),
		QuestionMessage:        q,
		inboundIP:              inboundIP,
		DNSUpstream:            config.Config.PrimaryDNSServer,
		EDNSClientSubnetPolicy: config.Config.EDNSClientSubnetPolicy,
		EDNSClientSubnetIP:     config.Config.EDNSClientSubnetIP,
		externalIP:             config.Config.ExternalIP,
		MinimumTTL:             config.Config.MinimumTTL,
	}

	o.IPUsed = o.getEDNSClientSubnetIP()

	return o
}

func (o *Outbound) ExchangeFromRemote(isEDNSClientSubnet bool) error {

	if  isEDNSClientSubnet {
		o.HandleEDNSClientSubnet()
	}
	c := new(dns.Client)
	c.Net = o.DNSUpstream.Protocol
	c.Timeout = time.Duration(c.Timeout) * time.Second
	temp, _, err := c.Exchange(o.QuestionMessage, o.DNSUpstream.Address)
	if err != nil {
		if err == dns.ErrTruncated {
			log.Warn("Maybe your primary dns server does not support edns client subnet")
		}
		return err
	}
	if temp == nil {
		err = errors.New("Response message is nil, maybe timeout")
		return err
	}
	o.ResponseMessage = temp
	return nil
}

func (o *Outbound) ExchangeFromLocal() bool {

	raw_ip := o.QuestionMessage.Question[0].Name
	ip := net.ParseIP(raw_ip[:len(raw_ip)-1])
	if ip.To4() != nil {
		a, _ := dns.NewRR(raw_ip + " IN A " + ip.String())
		o.ResponseMessage.Answer = append(o.ResponseMessage.Answer, a)
		o.ResponseMessage.SetReply(o.QuestionMessage)
		o.ResponseMessage.RecursionAvailable = true
		return true
	}
	return false
}

func (o *Outbound) HandleEDNSClientSubnet() {

	setEDNSClientSubnet(o.QuestionMessage, o.IPUsed)
}

func (o *Outbound) getEDNSClientSubnetIP() string{

	switch o.EDNSClientSubnetPolicy {
	case "custom":
		return o.EDNSClientSubnetIP
	case "auto":
		if !common.IsIPMatchList(net.ParseIP(o.inboundIP), config.Config.ReservedIPNetworkList, false) {
			return o.inboundIP
		} else {
			return o.externalIP
		}
	case "disable":
	}
	return ""
}

func (o *Outbound) HandleMinimumTTL() {

	if o.MinimumTTL > 0 {
		setMinimumTTL(o.ResponseMessage, uint32(o.MinimumTTL))
	}
}

func setMinimumTTL(m *dns.Msg, ttl uint32) {

	for _, a := range m.Answer {
		if a.Header().Ttl < ttl {
			a.Header().Ttl = ttl
		}
	}
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
