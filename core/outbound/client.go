// Package outbound implements multiple dns client and dispatcher for outbound connection.
package outbound

import (
	"math/rand"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
	"golang.org/x/net/proxy"
)

type Client struct {
	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	DNSUpstream           *DNSUpstream
	EDNSClientSubnetIP    string
	InboundIP             string
	ReservedIPNetworkList []*net.IPNet

	Hosts *hosts.Hosts
	Cache *cache.Cache
}

func NewClient(q *dns.Msg, u *DNSUpstream, ip string, h *hosts.Hosts, cache *cache.Cache) *Client {

	c := &Client{QuestionMessage: q, DNSUpstream: u, InboundIP: ip, Hosts: h, Cache: cache}
	c.getEDNSClientSubnetIP()
	c.ReservedIPNetworkList = getReservedIPNetworkList()
	return c
}

func (c *Client) getEDNSClientSubnetIP() {

	switch c.DNSUpstream.EDNSClientSubnet.Policy {
	case "auto":
		if !common.IsIPMatchList(net.ParseIP(c.InboundIP), c.ReservedIPNetworkList, false) {
			c.EDNSClientSubnetIP = c.InboundIP
		} else {
			c.EDNSClientSubnetIP = c.DNSUpstream.EDNSClientSubnet.ExternalIP
		}
	case "disable":
	}
}

func (c *Client) ExchangeFromRemote(isCache bool, isLog bool) {

	if c.ExchangeFromCache(isLog) {
		return
	}

	setEDNSClientSubnet(c.QuestionMessage, c.EDNSClientSubnetIP)

	var conn net.Conn
	if c.DNSUpstream.SOCKS5Address != "" {
		s, err := proxy.SOCKS5(c.DNSUpstream.Protocol, c.DNSUpstream.SOCKS5Address, nil, proxy.Direct)
		if err != nil {
			log.Warn("Get socks5 proxy dialer failed: ", err)
			return
		}
		conn, err = s.Dial(c.DNSUpstream.Protocol, c.DNSUpstream.Address)
		if err != nil {
			log.Warn("Dial DNS upstream with SOCKS5 proxy failed: ", err)
			return
		}
	} else {
		var err error
		if conn, err = net.Dial(c.DNSUpstream.Protocol, c.DNSUpstream.Address); err != nil {
			log.Warn("Dial DNS upstream failed: ", err)
			return
		}
	}

	dnsTimeout := time.Duration(c.DNSUpstream.Timeout) * time.Second / 3

	conn.SetDeadline(time.Now().Add(dnsTimeout))
	conn.SetReadDeadline(time.Now().Add(dnsTimeout))
	conn.SetWriteDeadline(time.Now().Add(dnsTimeout))

	dc := &dns.Conn{Conn: conn}
	defer dc.Close()
	err := dc.WriteMsg(c.QuestionMessage)
	if err != nil {
		log.Warn(c.DNSUpstream.Name + " Fail: Send question message failed")
		return
	}
	temp, err := dc.ReadMsg()

	if err != nil {
		if err == dns.ErrTruncated {
			log.Warn(c.DNSUpstream.Name + " Fail: Maybe this server does not support EDNS Client Subnet")
			return
		}
	}
	if temp == nil {
		log.Debug(c.DNSUpstream.Name + " Fail: Response message is nil, maybe timeout, please check your query or dns configuration")
		return
	}

	c.ResponseMessage = temp

	if isCache && c.Cache != nil {
		c.Cache.InsertMessage(cache.Key(c.QuestionMessage.Question[0], c.EDNSClientSubnetIP), c.ResponseMessage)
	}

	if isLog {
		c.logAnswer(false)
	}
}

func (c *Client) ExchangeFromCache(isLog bool) bool {

	if c.Cache == nil {
		return false
	}

	m := c.Cache.Hit(cache.Key(c.QuestionMessage.Question[0], c.EDNSClientSubnetIP), c.QuestionMessage.Id)
	if m != nil {
		log.Debug(c.DNSUpstream.Name + " Hit: " + cache.Key(c.QuestionMessage.Question[0], c.EDNSClientSubnetIP))
		c.ResponseMessage = m
		if isLog {
			c.logAnswer(false)
		}
		return true
	}

	return false
}

func (c *Client) ExchangeFromLocal() bool {
	raw_name := c.QuestionMessage.Question[0].Name

	if c.ExchangeFromHosts(raw_name) || c.ExchangeFromIP(raw_name) {
		return true
	}

	return false
}

func (c *Client) ExchangeFromHosts(raw_name string) bool {

	if c.Hosts == nil {
		return false
	}

	name := raw_name[:len(raw_name)-1]
	ipv4List, ipv6List := c.Hosts.Find(name)

	if c.QuestionMessage.Question[0].Qtype == dns.TypeA && len(ipv4List) > 0 {
		var rrl []dns.RR
		for _, ip := range ipv4List {
			a, _ := dns.NewRR(raw_name + " IN A " + ip.String())
			rrl = append(rrl, a)
		}
		c.setLocalResponseMessage(rrl)
		if c.ResponseMessage != nil {
			return true
		}
	} else if c.QuestionMessage.Question[0].Qtype == dns.TypeAAAA && len(ipv6List) > 0 {
		var rrl []dns.RR
		for _, ip := range ipv6List {
			aaaa, _ := dns.NewRR(raw_name + " IN AAAA " + ip.String())
			rrl = append(rrl, aaaa)
		}
		c.setLocalResponseMessage(rrl)
		if c.ResponseMessage != nil {
			return true
		}
	}

	return false
}

func (c *Client) ExchangeFromIP(raw_name string) bool {

	name := raw_name[:len(raw_name)-1]
	ip := net.ParseIP(name)
	if ip == nil {
		return false
	}
	if ip.To4() == nil && ip.To16() != nil && c.QuestionMessage.Question[0].Qtype == dns.TypeAAAA {
		aaaa, _ := dns.NewRR(raw_name + " IN AAAA " + ip.String())
		c.setLocalResponseMessage([]dns.RR{aaaa})
		return true
	} else if ip.To4() != nil && c.QuestionMessage.Question[0].Qtype == dns.TypeA {
		a, _ := dns.NewRR(raw_name + " IN A " + ip.String())
		c.setLocalResponseMessage([]dns.RR{a})
		return true
	}

	return false
}

func (c *Client) setLocalResponseMessage(rrl []dns.RR) {

	shuffleRRList := func(rrl []dns.RR) {
		rand.Seed(time.Now().UnixNano())
		for i := range rrl {
			j := rand.Intn(i + 1)
			rrl[i], rrl[j] = rrl[j], rrl[i]
		}
	}
	c.ResponseMessage = new(dns.Msg)
	for _, rr := range rrl {
		c.ResponseMessage.Answer = append(c.ResponseMessage.Answer, rr)
	}
	shuffleRRList(c.ResponseMessage.Answer)
	c.ResponseMessage.SetReply(c.QuestionMessage)
	c.ResponseMessage.RecursionAvailable = true
}

func (c *Client) logAnswer(isLocal bool) {

	for _, a := range c.ResponseMessage.Answer {
		var name string
		if isLocal {
			name = "Local"
		} else {
			name = c.DNSUpstream.Name
		}
		log.Debug(name + " Answer: " + a.String())
	}
}

func getReservedIPNetworkList() []*net.IPNet {

	ipnl := make([]*net.IPNet, 0)
	localCIDR := []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10"}
	for _, c := range localCIDR {
		_, ip_net, err := net.ParseCIDR(c)
		if err != nil {
			break
		}
		ipnl = append(ipnl, ip_net)
	}
	return ipnl
}
