package outbound

import (
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

	Hosts     *hosts.Hosts
	CachePool *cache.Cache
}

func NewClient(q *dns.Msg, u *DNSUpstream, ip string, h *hosts.Hosts, cp *cache.Cache) *Client {

	c := &Client{QuestionMessage: q, DNSUpstream: u, InboundIP: ip, Hosts: h, CachePool: cp}
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
	dc.WriteMsg(c.QuestionMessage)
	temp, err := dc.ReadMsg()
	dc.Close()

	if err != nil {
		if err == dns.ErrTruncated {
			log.Warn("Maybe your primary dns server does not support edns client subnet")
			return
		}
	}
	if temp == nil {
		log.Debug(c.DNSUpstream.Name + " Fail: Response message is nil, maybe timeout, please check your query or dns configuration")
		return
	}

	c.ResponseMessage = temp

	if isCache {
		c.CachePool.InsertMessage(cache.Key(c.QuestionMessage.Question[0], c.EDNSClientSubnetIP), c.ResponseMessage)
	}

	if isLog {
		c.logAnswer(false)
	}
}

func (c *Client) ExchangeFromCache(isLog bool) bool {

	if c.CachePool == nil {
		return false
	}

	m := c.CachePool.Hit(cache.Key(c.QuestionMessage.Question[0], c.EDNSClientSubnetIP), c.QuestionMessage.Id)
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

func (c *Client) ExchangeFromHosts(raw_name string) bool {

	if c.Hosts == nil {
		return false
	}

	name := raw_name[:len(raw_name)-1]
	ipl := c.Hosts.Find(name)

	if len(ipl) > 0 {
		for _, ip := range ipl {
			if c.QuestionMessage.Question[0].Qtype == dns.TypeA {
				a, _ := dns.NewRR(raw_name + " IN A " + ip.String())
				c.createResponseMessage(a)
				return true
			}
			if c.QuestionMessage.Question[0].Qtype == dns.TypeAAAA {
				aaaa, _ := dns.NewRR(raw_name + " IN AAAA " + ip.String())
				c.createResponseMessage(aaaa)
				return true
			}
		}
	}

	return false
}

func (c *Client) ExchangeFromIP(raw_name string) bool {

	name := raw_name[:len(raw_name)-1]
	ip := net.ParseIP(name)
	if ip.To4() != nil && c.QuestionMessage.Question[0].Qtype == dns.TypeA {
		a, _ := dns.NewRR(raw_name + " IN A " + ip.String())
		c.createResponseMessage(a)
		return true
	}
	if ip.To16() != nil && c.QuestionMessage.Question[0].Qtype == dns.TypeAAAA {
		aaaa, _ := dns.NewRR(raw_name + " IN AAAA " + ip.String())
		c.createResponseMessage(aaaa)
		return true
	}

	return false
}

func (c *Client) createResponseMessage(r dns.RR) {
	c.ResponseMessage = new(dns.Msg)
	c.ResponseMessage.Answer = append(c.ResponseMessage.Answer, r)
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

func (c *Client) ExchangeFromLocal() bool {
	raw_name := c.QuestionMessage.Question[0].Name

	if c.ExchangeFromHosts(raw_name) || c.ExchangeFromIP(raw_name) {
		return true
	}

	return false
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
