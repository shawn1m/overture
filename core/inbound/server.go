package inbound

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/outbound"
	"net"
	"os"
	"sync"
)

type Server struct {
	BindAddress string `json:"BindAddress"`

	Dispatcher     *outbound.Dispatcher
	OnlyPrimaryDNS bool

	MinimumTTL  int
	RejectQtype []uint16

	Hosts     *hosts.Hosts
	CachePool *cache.Cache
}

func (s *Server) Run() {

	mux := dns.NewServeMux()
	mux.Handle(".", s)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	log.Info("Start overture on " + s.BindAddress)

	for _, p := range [2]string{"tcp", "udp"} {
		go func(p string) {
			err := dns.ListenAndServe(s.BindAddress, p, mux)
			if err != nil {
				log.Fatal("Listen "+p+" failed: ", err)
				os.Exit(1)
			}
		}(p)
	}

	wg.Wait()
}

func (s *Server) ServeDNS(w dns.ResponseWriter, q *dns.Msg) {

	inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	cb := outbound.NewClientBundle(q, s.Dispatcher.PrimaryDNS, inboundIP, s.Hosts, s.CachePool)
	s.Dispatcher.ClientBundle = cb

	log.Debug("Question: " + cb.QuestionMessage.Question[0].String())

	for _, qt := range s.RejectQtype {
		if isQuestionType(q, qt) {
			return
		}
	}

	if ok := cb.ExchangeFromLocal(); ok {
		if cb.ResponseMessage != nil {
			w.WriteMsg(cb.ResponseMessage)
			return
		}
	}

	func() {
		if s.OnlyPrimaryDNS {
			cb.ExchangeFromRemote(true, true)
		} else {
			if ok := s.Dispatcher.ExchangeForIPv6() || s.Dispatcher.ExchangeForDomain(); ok {
				return
			}

			cb.ExchangeFromRemote(false, true)
			s.Dispatcher.ExchangeForPrimaryDNSResponse()
		}
	}()

	if s.MinimumTTL > 0 {
		setMinimumTTL(cb.ResponseMessage, uint32(s.MinimumTTL))
	}

	if cb.ResponseMessage != nil {
		w.WriteMsg(cb.ResponseMessage)
	}
}

func isQuestionType(q *dns.Msg, qt uint16) bool { return q.Question[0].Qtype == qt }

func setMinimumTTL(m *dns.Msg, ttl uint32) {

	for _, a := range m.Answer {
		if a.Header().Ttl < ttl {
			a.Header().Ttl = ttl
		}
	}
}
