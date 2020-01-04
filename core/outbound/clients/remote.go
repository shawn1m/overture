/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package clients

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
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
	common.SetEDNSClientSubnet(c.questionMessage, c.ednsClientSubnetIP,
		c.dnsUpstream.EDNSClientSubnet.NoCookie)
	c.ednsClientSubnetIP = common.GetEDNSClientSubnetIP(c.questionMessage)

	if c.responseMessage != nil {
		return c.responseMessage
	}

	var conn net.Conn = nil
	var err error
	if c.dnsUpstream.SOCKS5Address != "" {
		if conn, err = c.createSocks5Conn(); err != nil {
			return nil
		}
	}

	var temp *dns.Msg
	switch c.dnsUpstream.Protocol {
	case "udp":
		temp, err = c.ExchangeByUDP(conn)
	case "tcp":
		temp, err = c.ExchangeByTCP(conn)
	case "tcp-tls":
		temp, err = c.ExchangeByTLS(conn)
	case "https":
		temp, err = c.ExchangeByHTTPS(conn)
	}

	if err != nil {
		log.Debugf("%s Fail: %s", c.dnsUpstream.Name, err)
		return nil
	}
	if temp == nil {
		log.Debugf("%s Fail: Response message returned nil, maybe timeout? Please check your query or DNS configuration")
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
		log.Debugf("Answer from %s: %s", name, a.String())
	}
}

func (c *RemoteClient) createSocks5Conn() (conn net.Conn, err error) {
	socksAddress, err := ExtractSocksAddress(c.dnsUpstream.SOCKS5Address)
	if err != nil {
		return nil, err
	}
	network := ToNetwork(c.dnsUpstream.Protocol)
	s, err := proxy.SOCKS5(network, socksAddress, nil, proxy.Direct)
	if err != nil {
		log.Warnf("Failed to connect to SOCKS5 proxy: %s", err)
		return nil, err
	}
	host, port, err := ExtractDNSAddress(c.dnsUpstream.Address, c.dnsUpstream.Protocol)
	if err != nil {
		return nil, err
	}
	address := net.JoinHostPort(host, port)
	conn, err = s.Dial(network, address)
	if err != nil {
		log.Warnf("Failed to connect to upstream via SOCKS5 proxy: %s", err)
		return nil, err
	}
	return conn, err
}

func (c *RemoteClient) exchangeByDNSClient(conn net.Conn) (msg *dns.Msg, err error) {
	if conn == nil {
		network := ToNetwork(c.dnsUpstream.Protocol)
		host, port, err := ExtractDNSAddress(c.dnsUpstream.Address, c.dnsUpstream.Protocol)
		if err != nil {
			return nil, err
		}
		address := net.JoinHostPort(host, port)
		if conn, err = net.Dial(network, address); err != nil {
			log.Warnf("Failed to connect to DNS upstream: %s", err)
			return nil, err
		}
	}
	c.setTimeout(conn)
	dc := &dns.Conn{Conn: conn, UDPSize: 65535}
	defer dc.Close()
	err = dc.WriteMsg(c.questionMessage)
	if err != nil {
		log.Warnf("%s Fail: Send question message failed", c.dnsUpstream.Name)
		return nil, err
	}
	return dc.ReadMsg()
}

// ExchangeByUDP send dns record by udp protocol
func (c *RemoteClient) ExchangeByUDP(conn net.Conn) (*dns.Msg, error) {
	return c.exchangeByDNSClient(conn)
}

// ExchangeByTCP send dns record by tcp protocol
func (c *RemoteClient) ExchangeByTCP(conn net.Conn) (*dns.Msg, error) {
	return c.exchangeByDNSClient(conn)
}

// ExchangeByTLS send dns record by tcp-tls protocol
func (c *RemoteClient) ExchangeByTLS(conn net.Conn) (msg *dns.Msg, err error) {
	host, port, ip := ExtractTLSDNSAddress(c.dnsUpstream.Address)
	var address string
	if len(ip) > 0 {
		address = net.JoinHostPort(ip, port)
	} else {
		address = net.JoinHostPort(host, port)
	}

	conf := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}
	if conn != nil {
		// crate tls client use the existing connection
		conn = tls.Client(conn, conf)
	} else {
		if conn, err = tls.Dial("tcp", address, conf); err != nil {
			log.Warnf("Failed to connect to DNS-over-TLS upstream: %s", err)
			return nil, err
		}
	}
	c.setTimeout(conn)
	return c.exchangeByDNSClient(conn)
}

// ExchangeByHTTPS send dns record by https protocol
func (c *RemoteClient) ExchangeByHTTPS(conn net.Conn) (*dns.Msg, error) {
	if conn == nil {
		host, port, err := ExtractHTTPSAddress(c.dnsUpstream.Address)
		if err != nil {
			return nil, err
		}
		address := net.JoinHostPort(host, port)
		conn, err = net.Dial("tcp", address)
		if err != nil {
			log.Warnf("Fail connect to dns server %s", address)
		}
	}
	c.setTimeout(conn)
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return conn, nil
			},
		},
	}
	defer client.CloseIdleConnections()
	request, err := c.questionMessage.Pack()
	resp, err := client.Post(c.dnsUpstream.Address, "application/dns-message",
		bytes.NewBuffer(request))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	msg := new(dns.Msg)
	err = msg.Unpack(data)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *RemoteClient) setTimeout(conn net.Conn) {
	dnsTimeout := time.Duration(c.dnsUpstream.Timeout) * time.Second / 3
	conn.SetDeadline(time.Now().Add(dnsTimeout))
	conn.SetReadDeadline(time.Now().Add(dnsTimeout))
	conn.SetWriteDeadline(time.Now().Add(dnsTimeout))
}
