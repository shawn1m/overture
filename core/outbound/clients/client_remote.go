/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package clients

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
)

type RemoteClient struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	dnsUpstream        *common.DNSUpstream
	ednsClientSubnetIP string
	inboundIP          string

	cache *cache.Cache
}

func NewClient(q *dns.Msg, u *common.DNSUpstream, ip string, cache *cache.Cache) *RemoteClient {

	c := &RemoteClient{questionMessage: q.Copy(), dnsUpstream: u, inboundIP: ip, cache: cache}
	c.getEDNSClientSubnetIP()

	return c
}

func (c *RemoteClient) getEDNSClientSubnetIP() {

	switch c.dnsUpstream.EDNSClientSubnet.Policy {
	case "auto":
		if !common.IsIPMatchList(net.ParseIP(c.inboundIP), common.ReservedIPNetworkList, false, "") {
			c.ednsClientSubnetIP = c.inboundIP
		} else {
			c.ednsClientSubnetIP = c.dnsUpstream.EDNSClientSubnet.ExternalIP
		}
	case "manual":
		if c.dnsUpstream.EDNSClientSubnet.ExternalIP != "" &&
			!common.IsIPMatchList(net.ParseIP(c.dnsUpstream.EDNSClientSubnet.ExternalIP), common.ReservedIPNetworkList, false, "") {
			c.ednsClientSubnetIP = c.dnsUpstream.EDNSClientSubnet.ExternalIP
			return
		}
	case "disable":
	}
}

func (c *RemoteClient) ExchangeFromCache() *dns.Msg {
	cacheClient := NewCacheClient(c.questionMessage, c.ednsClientSubnetIP, c.cache)
	c.responseMessage = cacheClient.Exchange()
	if c.responseMessage != nil {
		return c.responseMessage
	}
	return nil
}

func (c *RemoteClient) Exchange(isLog bool) *dns.Msg {

	common.SetEDNSClientSubnet(c.questionMessage, c.ednsClientSubnetIP, c.dnsUpstream.EDNSClientSubnet.NoCookie)
	c.ednsClientSubnetIP = common.GetEDNSClientSubnetIP(c.questionMessage)

	if c.responseMessage != nil {
		return c.responseMessage
	}

	var conn net.Conn
	if c.dnsUpstream.SOCKS5Address != "" {
		s, err := proxy.SOCKS5(c.dnsUpstream.Protocol, c.dnsUpstream.SOCKS5Address, nil, proxy.Direct)
		if err != nil {
			log.Warn("Get socks5 proxy dialer failed: ", err)
			return nil
		}
		conn, err = s.Dial(c.dnsUpstream.Protocol, c.dnsUpstream.Address)
		if err != nil {
			log.Warn("Dial DNS upstream with SOCKS5 proxy failed: ", err)
			return nil
		}
	} else if c.dnsUpstream.Protocol == "tcp-tls" {
		var err error
		conf := &tls.Config{
			InsecureSkipVerify: false,
		}
		s := strings.Split(c.dnsUpstream.Address, "@")
		if len(s) == 2 {
			var servername, port string
			if servername, port, err = net.SplitHostPort(s[0]); err != nil {
				log.Warn("DNS-over-TLS servername:port@serverAddress config failed: ", err)
				return nil
			}
			conf.ServerName = servername
			c.dnsUpstream.Address = s[1] + ":" + port
		}
		if conn, err = tls.Dial("tcp", c.dnsUpstream.Address, conf); err != nil {
			log.Warn("Dial DNS-over-TLS upstream failed: ", err)
			return nil
		}
	} else {
		var err error
		if conn, err = net.Dial(c.dnsUpstream.Protocol, c.dnsUpstream.Address); err != nil {
			log.Warn("Dial DNS upstream failed: ", err)
			return nil
		}
	}

	dnsTimeout := time.Duration(c.dnsUpstream.Timeout) * time.Second / 3

	conn.SetDeadline(time.Now().Add(dnsTimeout))
	conn.SetReadDeadline(time.Now().Add(dnsTimeout))
	conn.SetWriteDeadline(time.Now().Add(dnsTimeout))

	dc := &dns.Conn{Conn: conn}
	defer dc.Close()
	err := dc.WriteMsg(c.questionMessage)
	if err != nil {
		log.Warn(c.dnsUpstream.Name + " Fail: Send question message failed")
		return nil
	}
	temp, err := dc.ReadMsg()

	if err != nil {
		log.Debug(c.dnsUpstream.Name+" Fail: ", err)
		return nil
	}
	if temp == nil {
		log.Debug(c.dnsUpstream.Name + " Fail: Response message is nil, maybe timeout, please check your query or dns configuration")
		return nil
	}

	c.responseMessage = temp

	if isLog {
		c.logAnswer("")
	}

	return c.responseMessage
}

func (c *RemoteClient) logAnswer(indicator string) {

	for _, a := range c.responseMessage.Answer {
		var name string
		if indicator != "" {
			name = indicator
		} else {
			name = c.dnsUpstream.Name
		}
		log.Debug("Answer from " + name + ": " + a.String())
	}
}
