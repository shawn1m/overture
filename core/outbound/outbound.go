package outbound

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/cache"
	"github.com/holyshawn/overture/core/common"
	"github.com/holyshawn/overture/core/config"
	"github.com/miekg/dns"
)

type Outbound struct {
	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	DNSUpstream        *config.DNSUpstream
	MinimumTTL         int
	EDNSClientSubnetIP string

	inboundIP string
}

func NewOutbound(q *dns.Msg, d *config.DNSUpstream, inboundIP string) *Outbound {

	o := &Outbound{
		QuestionMessage: q,

		DNSUpstream: d,
		MinimumTTL:  config.Config.MinimumTTL,

		inboundIP: inboundIP,
	}

	o.EDNSClientSubnetIP = o.getEDNSClientSubnetIP()

	return o
}

func (o *Outbound) ExchangeFromRemote(IsCache bool) {

	m := config.Config.CachePool.Hit(cache.Key(o.QuestionMessage.Question[0], o.EDNSClientSubnetIP), o.QuestionMessage.Id)
	if m != nil {
		log.Debug(o.DNSUpstream.Name + " Hit: " + cache.Key(o.QuestionMessage.Question[0], o.EDNSClientSubnetIP))
		o.ResponseMessage = m
		o.LogAnswer(false)
		return
	}
	m = config.Config.CachePool.Hit(cache.Key(o.QuestionMessage.Question[0], ""), o.QuestionMessage.Id)
	if m != nil {
		o.ResponseMessage = m
		log.Debug(o.DNSUpstream.Name + " Hit: " + cache.Key(o.QuestionMessage.Question[0], ""))
		o.LogAnswer(false)
		return
	}

	o.HandleEDNSClientSubnet()

	c := new(dns.Client)
	c.Net = o.DNSUpstream.Protocol
	c.Timeout = time.Duration(o.DNSUpstream.Timeout) * time.Second

	temp, _, err := c.Exchange(o.QuestionMessage, o.DNSUpstream.Address)
	if err != nil {
		if err == dns.ErrTruncated {
			log.Warn("Maybe your primary dns server does not support edns client subnet")
			return
		}
	}
	if temp == nil {
		log.Debug(o.DNSUpstream.Name + " Fail: Response message is nil, maybe timeout")
		return
	}

	o.ResponseMessage = temp

	o.HandleMinimumTTL()

	if IsCache {
		config.Config.CachePool.InsertMessage(cache.Key(o.QuestionMessage.Question[0], o.EDNSClientSubnetIP), o.ResponseMessage)
	}

	o.LogAnswer(false)
}

func (o *Outbound) LogAnswer(isLocal bool) {

	for _, a := range o.ResponseMessage.Answer {
		var name string
		if isLocal{
			name = "Local"
		}else {
			name = o.DNSUpstream.Name
		}
		log.Debug(name + " Answer: " + a.String())
	}
}

func (o *Outbound) ExchangeFromLocal() bool {

	raw_name := o.QuestionMessage.Question[0].Name
	name := raw_name[:len(raw_name)-1]

	if config.Config.Hosts != nil {
		ipl, err := config.Config.Hosts.FindHosts(name)

		if err == nil && len(ipl) > 0 {
			for _, ip := range ipl {
				a, _ := dns.NewRR(raw_name + " IN A " + ip.String())
				o.ResponseMessage = new(dns.Msg)
				o.ResponseMessage.Answer = append(o.ResponseMessage.Answer, a)
			}
			o.ResponseMessage.SetReply(o.QuestionMessage)
			o.ResponseMessage.RecursionAvailable = true
			return true
		}
	}

	ip := net.ParseIP(name)
	if ip.To4() != nil {
		a, _ := dns.NewRR(raw_name + " IN A " + ip.String())
		o.ResponseMessage = new(dns.Msg)
		o.ResponseMessage.Answer = append(o.ResponseMessage.Answer, a)
		o.ResponseMessage.SetReply(o.QuestionMessage)
		o.ResponseMessage.RecursionAvailable = true
		return true
	}

	return false
}

func (o *Outbound) HandleEDNSClientSubnet() {

	setEDNSClientSubnet(o.QuestionMessage, o.EDNSClientSubnetIP)
}

func (o *Outbound) getEDNSClientSubnetIP() string {

	switch o.DNSUpstream.EDNSClientSubnet.Policy {
	case "custom":
		return o.DNSUpstream.EDNSClientSubnet.CustomIP
	case "auto":
		if !common.IsIPMatchList(net.ParseIP(o.inboundIP), config.Config.ReservedIPNetworkList, false) {
			return o.inboundIP
		} else {
			return o.DNSUpstream.EDNSClientSubnet.ExternalIP
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
