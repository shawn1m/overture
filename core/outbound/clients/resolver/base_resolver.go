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
	Init() error
}

type BaseResolver struct {
	dnsUpstream *common.DNSUpstream
}

func (r *BaseResolver) Exchange(*dns.Msg) (*dns.Msg, error) {
	return nil, nil
}

func (r *BaseResolver) ExchangeByBaseConn(q *dns.Msg) (*dns.Msg, error) {
	conn, err := r.CreateBaseConn()
	if err != nil {
		return nil, err
	}
	r.setTimeout(conn)
	dc := &dns.Conn{Conn: conn, UDPSize: 65535}
	defer dc.Close()
	err = dc.WriteMsg(q)
	if err != nil {
		return nil, err
	}
	return dc.ReadMsg()
}

func (r *BaseResolver) Init() error {
	if r.dnsUpstream.TCPPoolConfig.Enable {
		if r.dnsUpstream.TCPPoolConfig.IdleTimeout != 0 {
			IdleTimeout = time.Duration(r.dnsUpstream.TCPPoolConfig.IdleTimeout) * time.Second
		}
		if r.dnsUpstream.TCPPoolConfig.MaxCapacity != 0 {
			MaxCapacity = r.dnsUpstream.TCPPoolConfig.MaxCapacity
		}
		if r.dnsUpstream.TCPPoolConfig.InitialCapacity != 0 {
			InitialCapacity = r.dnsUpstream.TCPPoolConfig.InitialCapacity
		}
	}
	return nil
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
		log.Errorf("Create resolver for %s failed", u.Name)
		return nil
	}
	err := resolver.Init()
	if err != nil {
		log.Errorf("Init resolver for %s failed", u.Name)
	} else {
		log.Debugf("Init resolver for %s succeed", u.Name)
	}
	return resolver
}

func (r *BaseResolver) CreateBaseConn() (net.Conn, error) {
	dialer := net.Dial
	if r.dnsUpstream.SOCKS5Address != "" {
		socksAddress, err := ExtractSocksAddress(r.dnsUpstream.SOCKS5Address)
		if err != nil {
			return nil, err
		}
		network := ToNetwork(r.dnsUpstream.Protocol)
		s, err := proxy.SOCKS5(network, socksAddress, nil, proxy.Direct)
		if err != nil {
			log.Warnf("Failed to connect to SOCKS5 proxy: %s", err)
			return nil, err
		}
		dialer = s.Dial
	}

	network := ToNetwork(r.dnsUpstream.Protocol)
	host, port, err := ExtractDNSAddress(r.dnsUpstream.Address, r.dnsUpstream.Protocol)
	if err != nil {
		return nil, err
	}
	address := net.JoinHostPort(host, port)
	log.Debugf("Creating new connection to %s:%s", host, port)
	var conn net.Conn
	if conn, err = dialer(network, address); err != nil {
		log.Warnf("Failed to connect to DNS upstream: %s", err)
		return nil, err
	}

	// the Timeout setting is now moved to each resolver to support pool's idle timeout
	// r.setTimeout(conn)
	return conn, err
}

var InitialCapacity = 0
var IdleTimeout = 30 * time.Second
var MaxCapacity = 15

func (r *BaseResolver) setTimeout(conn net.Conn) {
	dnsTimeout := time.Duration(r.dnsUpstream.Timeout) * time.Second / 3
	conn.SetDeadline(time.Now().Add(dnsTimeout))
	conn.SetReadDeadline(time.Now().Add(dnsTimeout))
	conn.SetWriteDeadline(time.Now().Add(dnsTimeout))
}

func (r *BaseResolver) setIdleTimeout(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(IdleTimeout))
	conn.SetReadDeadline(time.Now().Add(IdleTimeout))
	conn.SetWriteDeadline(time.Now().Add(IdleTimeout))
}

func (r *BaseResolver) createConnectionPool(connCreate func() (interface{}, error), connClose func(interface{}) error) (pool.Pool, error) {
	poolConfig := &pool.Config{
		InitialCap: InitialCapacity,
		MaxCap:     MaxCapacity,
		Factory:    connCreate,
		Close:      connClose,
		//Ping:       ping,
		IdleTimeout: IdleTimeout,
	}
	return pool.NewChannelPool(poolConfig)
}

func (r *BaseResolver) exchangeByDNSClient(q *dns.Msg, conn net.Conn) (msg *dns.Msg, err error) {
	if conn == nil {
		log.Fatal("Conn not initialized for exchangeByDNSClient")
		return nil, err
	}

	dc := &dns.Conn{Conn: conn, UDPSize: 65535}
	err = dc.WriteMsg(q)
	if err != nil {
		log.Warnf("%s Fail: Send question message failed", r.dnsUpstream.Name)
		return nil, err
	}
	return dc.ReadMsg()
}
