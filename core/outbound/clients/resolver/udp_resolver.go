package resolver

import (
	"github.com/miekg/dns"
)

type UDPResolver struct {
	BaseResolver
}

func (r *UDPResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
	return r.BaseResolver.Exchange(q)
}

func (r *UDPResolver) Init() error {
	err := r.BaseResolver.Init()
	if err != nil {
		return err
	}
	return nil
}
