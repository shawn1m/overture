package resolver

import (
	"github.com/miekg/dns"
)

type UDPResolver struct {
	BaseResolver
}

func (r *UDPResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
	conn, err := r.CreateBaseConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	r.setTimeout(conn)
	ret, err := r.exchangeByDNSClient(q, conn)
	return ret, err
}

func (r *UDPResolver) Init() error {
	err := r.BaseResolver.Init()
	if err != nil {
		return err
	}
	return nil
}
