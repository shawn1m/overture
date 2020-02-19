/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */
package resolver

import (
	"github.com/miekg/dns"
	"github.com/silenceper/pool"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"net"
	"time"

	"github.com/shawn1m/overture/core/common"
)

type Resolver interface {
	Exchange(*dns.Msg) (*dns.Msg, error)
	Init()
}

type BaseResolver struct {
	dnsUpstream *common.DNSUpstream
}

func (r *BaseResolver) Exchange(*dns.Msg) (*dns.Msg, error) {
	return nil, nil
}

func (r *BaseResolver) Init() {

}

func NewResolver(u *common.DNSUpstream) Resolver {
	var resolver Resolver
	switch u.Protocol {
	case "udp":
		resolver = &UDPResolver{BaseResolver: BaseResolver{u}}
	case "tcp":
		resolver = &TCPResolver{BaseResolver: BaseResolver{u}}
	case "tcp-tls":
		resolver = &TCPTLSResolver{BaseResolver: BaseResolver{u}}
	case "https":
		resolver = &HTTPSResolver{BaseResolver: BaseResolver{u}}
	default:
		log.Fatalf("Unsupported protocol: %s", u.Protocol)
		return nil
	}
	resolver.Init()
	return resolver
}

func (c *BaseResolver) CreateBaseConn() (conn net.Conn, err error) {
	dialer := net.Dial
	if c.dnsUpstream.SOCKS5Address != "" {
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
		dialer = s.Dial
	}

	network := ToNetwork(c.dnsUpstream.Protocol)
	host, port, err := ExtractDNSAddress(c.dnsUpstream.Address, c.dnsUpstream.Protocol)
	if err != nil {
		return nil, err
	}
	address := net.JoinHostPort(host, port)
	log.Debugf("Creating new connection to %s:%s", host, port)
	if conn, err = dialer(network, address); err != nil {
		log.Warnf("Failed to connect to DNS upstream: %s", err)
		return nil, err
	}

	// the Timeout setting is now moved to each resolver to support pool's idle timeout
	// c.setTimeout(conn)
	return conn, err
}

var InitialCapacity = 0
var IdleTimeout = 30 * time.Second
var MaxCapacity = 15

func (c *BaseResolver) setTimeout(conn net.Conn) {
	dnsTimeout := time.Duration(c.dnsUpstream.Timeout) * time.Second / 3
	conn.SetDeadline(time.Now().Add(dnsTimeout))
	conn.SetReadDeadline(time.Now().Add(dnsTimeout))
	conn.SetWriteDeadline(time.Now().Add(dnsTimeout))
}

func (c *BaseResolver) setIdleTimeout(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(IdleTimeout))
	conn.SetReadDeadline(time.Now().Add(IdleTimeout))
	conn.SetWriteDeadline(time.Now().Add(IdleTimeout))
}

func (c *BaseResolver) createConnectionPool(connCreate func() (interface{}, error), connClose func(interface{}) error) pool.Pool {
	poolConfig := &pool.Config{
		InitialCap: InitialCapacity,
		MaxCap:     MaxCapacity,
		Factory:    connCreate,
		Close:      connClose,
		//Ping:       ping,
		IdleTimeout: IdleTimeout,
	}
	ret, _ := pool.NewChannelPool(poolConfig)
	return ret
}

func (c *BaseResolver) exchangeByDNSClient(q *dns.Msg, conn net.Conn) (msg *dns.Msg, err error) {
	if conn == nil {
		log.Fatal("Conn not initialized for exchangeByDNSClient")
		return nil, err
	}

	dc := &dns.Conn{Conn: conn, UDPSize: 65535}
	err = dc.WriteMsg(q)
	if err != nil {
		log.Warnf("%s Fail: Send question message failed", c.dnsUpstream.Name)
		return nil, err
	}
	return dc.ReadMsg()
}
