package outbound

import (
	"errors"
	"net"
	"time"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/common"
)

type Outbound struct {
	ResponseMessage        *dns.Msg
	QuestionMessage        *dns.Msg
	InboundIP              string
	DomainNameServer       *config.DNSUpstream
	EDNSClientSubnetPolicy string
	EDNSClientSubnetIP     string
	ExternalIP             string
	MinimumTTL             int
}

func NewOutbound(q *dns.Msg, inboundIP string, d *config.DNSUpstream) *Outbound {

	return &Outbound{
		ResponseMessage: new(dns.Msg),
		QuestionMessage: q,
		InboundIP: inboundIP,
		DomainNameServer: d,
		EDNSClientSubnetPolicy: config.Config.EDNSClientSubnetPolicy,
		EDNSClientSubnetIP: config.Config.EDNSClientSubnetIP,
		ExternalIP: config.Config.ExternalIP,
		MinimumTTL: config.Config.MinimumTTL,
	}
}

func (o *Outbound) ExchangeFromRemote() error {

	if reflect.DeepEqual(o.DomainNameServer, config.Config.PrimaryDNSServer) {
		o.HandleEDNSsClientSubnet()
	}
	c := new(dns.Client)
	c.Net = o.DomainNameServer.Protocol
	c.Timeout = time.Duration(c.Timeout) * time.Second
	temp, _, err := c.Exchange(o.QuestionMessage, o.DomainNameServer.Address)
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

func (o *Outbound) HandleEDNSsClientSubnet() {

	switch o.EDNSClientSubnetPolicy {
	case "custom":
		setEDNSClientSubnet(o.QuestionMessage, o.EDNSClientSubnetIP)
	case "auto":
		if !common.IsIPMatchList(net.ParseIP(o.InboundIP), config.Config.ReservedIPNetworkList, false) {
			setEDNSClientSubnet(o.QuestionMessage, o.InboundIP)
		} else {
			setEDNSClientSubnet(o.QuestionMessage, o.ExternalIP)
		}
	case "disable":
	}
}

func (o *Outbound) HandleMinimumTTL() {

	if o.MinimumTTL > 0 {
		setMinimumTTL(o.ResponseMessage, uint32(o.MinimumTTL))
	}
}

func setMinimumTTL(m *dns.Msg, ttl uint32) {

	for _, answer := range m.Answer {
		if answer.Header().Ttl < ttl {
			answer.Header().Ttl = ttl
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

	es := isEDNSClientSubnet(o)
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

func isEDNSClientSubnet(o *dns.OPT) *dns.EDNS0_SUBNET {
	for _, s := range o.Option {
		switch e := s.(type) {
		case *dns.EDNS0_SUBNET:
			return e
		}
	}
	return nil
}