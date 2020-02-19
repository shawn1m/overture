package resolver

import (
	"github.com/miekg/dns"
)

type UDPResolver struct {
	BaseResolver
}

func (r *UDPResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
	// we don't need to pooling for UDP sockets
	conn, err := r.CreateBaseConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	r.setTimeout(conn)
	ret, err := r.exchangeByDNSClient(q, conn)
	return ret, err
}

func (r *UDPResolver) Init() {

}
