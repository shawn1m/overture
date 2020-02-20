package resolver

import (
	"net"

	"github.com/miekg/dns"
	"github.com/silenceper/pool"
	log "github.com/sirupsen/logrus"
)

type TCPResolver struct {
	BaseResolver
	poolConn pool.Pool
}

func (r *TCPResolver) Exchange(q *dns.Msg) (*dns.Msg, error) {
	if r.dnsUpstream.TCPPoolConfig.Enable {
		return r.BaseResolver.exchangeByPool(q, r.poolConn)
	} else {
		return r.BaseResolver.Exchange(q)
	}
}

func (r *TCPResolver) Init() error {
	err := r.BaseResolver.Init()
	if err != nil {
		return err
	}
	if r.dnsUpstream.TCPPoolConfig.Enable {
		r.poolConn, err = r.createConnectionPool(
			func() (interface{}, error) { return r.CreateBaseConn() },
			func(v interface{}) error { return v.(net.Conn).Close() })
		if err != nil {
			log.Debugf("Set %s pool's IdleTimeout to %d, InitialCapacity to %d, MaxCapacity to %d", r.dnsUpstream.Name, r.dnsUpstream.TCPPoolConfig.IdleTimeout, r.dnsUpstream.TCPPoolConfig.InitialCapacity, r.dnsUpstream.TCPPoolConfig.MaxCapacity)
		}
	} else {
		return nil
	}
	return err
}
