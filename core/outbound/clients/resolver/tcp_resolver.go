package resolver

import (
	"github.com/miekg/dns"
	"github.com/silenceper/pool"
	"net"
)

type TCPResolver struct {
	BaseResolver
	connpool pool.Pool
}

func (r *TCPResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
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

func (r *TCPResolver) Init() {
	r.connpool = r.createConnectionPool(
		func() (interface{}, error) { return r.CreateBaseConn() },
		func(v interface{}) error { return v.(net.Conn).Close() })
}
