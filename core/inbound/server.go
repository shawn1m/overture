// Package inbound implements dns server for inbound connection.
package inbound

import (
	"net"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/hosts"
	"github.com/shawn1m/overture/core/outbound"
	"github.com/ulule/deepcopier"
)

type Server struct {
	BindAddress string

	Dispatcher *outbound.Dispatcher

	MinimumTTL  int
	RejectQtype []uint16

	Hosts *hosts.Hosts
	Cache *cache.Cache
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
	qt := &dns.Msg{}
	deepcopier.Copy(q).To(qt)
	s.Dispatcher.PrimaryClientBundle = outbound.NewClientBundle(q, s.Dispatcher.PrimaryDNS, inboundIP, s.Hosts, s.Cache)
	s.Dispatcher.AlternativeClientBundle = outbound.NewClientBundle(qt, s.Dispatcher.AlternativeDNS, inboundIP, s.Hosts, s.Cache)

	log.Debug("Question: " + q.Question[0].String())

	for _, qt := range s.RejectQtype {
		if isQuestionType(q, qt) {
			return
		}
	}

	s.Dispatcher.Exchange()

	cb := s.Dispatcher.ActiveClientBundle

	if cb.ResponseMessage != nil {
		if s.MinimumTTL > 0 {
			setMinimumTTL(cb.ResponseMessage, uint32(s.MinimumTTL))
		}
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
