package resolver

import (
	"crypto/tls"
	"github.com/miekg/dns"
	"github.com/silenceper/pool"
	"net"
)

type TCPTLSResolver struct {
	BaseResolver

	connpool pool.Pool
}

func (r *TCPTLSResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
	_conn, err := r.connpool.Get()
	if err != nil {
		return nil, err
	}
	conn := _conn.(net.Conn)
	r.setTimeout(conn)
	ret, err := r.exchangeByDNSClient(q, conn)
	if err != nil {
		r.connpool.Close(conn)
	} else {
		r.setIdleTimeout(conn)
		r.connpool.Put(conn)
	}
	return ret, err
}

func (r *TCPTLSResolver) createTlsConn() (conn net.Conn, err error) {
	conn, err = r.CreateBaseConn()
	if conn == nil {
		return nil, err
	}
	host, _, _ := ExtractTLSDNSAddress(r.dnsUpstream.Address)
	conf := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}
	conn = tls.Client(conn, conf)
	//r.setTimeout(conn)
	return conn, nil
}

func (r *TCPTLSResolver) Init() {
	r.connpool = r.createConnectionPool(
		func() (interface{}, error) { return r.createTlsConn() },
		func(v interface{}) error { return v.(net.Conn).Close() })
}
