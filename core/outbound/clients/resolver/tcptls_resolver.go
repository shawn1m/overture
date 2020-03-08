package resolver

import (
	"crypto/tls"
	"net"

	"github.com/miekg/dns"
	"github.com/silenceper/pool"
	log "github.com/sirupsen/logrus"
)

type TCPTLSResolver struct {
	BaseResolver
	poolConn pool.Pool
}

func (r *TCPTLSResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
	if r.dnsUpstream.TCPPoolConfig.Enable {
		return r.BaseResolver.exchangeByPool(q, r.poolConn)
	} else {
		conn, err := r.createTlsConn()
		if err != nil {
			log.Warnf("createTlsConn failed: %s", err)
			return nil, err
		}
		defer conn.Close()
		return r.exchangeByConnWithoutClose(q, conn)
	}
}

func (r *TCPTLSResolver) createTlsConn() (conn net.Conn, err error) {
	conn, err = r.CreateBaseConn()
	if err != nil {
		return nil, err
	}
	host, err := ExtractTLSDNSHostName(r.dnsUpstream.Address)
	if err != nil {
		return nil, err
	}
	conf := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}
	conn = tls.Client(conn, conf)

	return conn, nil
}

func (r *TCPTLSResolver) Init() error {
	err := r.BaseResolver.Init()
	if err != nil {
		return err
	}
	if r.dnsUpstream.TCPPoolConfig.Enable {
		r.poolConn, err = r.createConnectionPool(
			func() (interface{}, error) { return r.createTlsConn() },
			func(v interface{}) error { return v.(net.Conn).Close() })
		if err != nil {
			log.Debugf("Set %s pool's IdleTimeout to %d, InitialCapacity to %d, MaxCapacity to %d", r.dnsUpstream.Name, r.dnsUpstream.TCPPoolConfig.IdleTimeout, r.dnsUpstream.TCPPoolConfig.InitialCapacity, r.dnsUpstream.TCPPoolConfig.MaxCapacity)
		}
	} else {
		return nil
	}
	return err
}
