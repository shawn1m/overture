package inbound

import (
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/common"
	"github.com/holyshawn/overture/core/outbound"
	"github.com/holyshawn/overture/core/switcher"
	"github.com/holyshawn/overture/core/cache"
	"github.com/miekg/dns"

	"github.com/holyshawn/overture/core/config"
)

func InitServer(addr string) {
	handler := dns.NewServeMux()
	handler.HandleFunc(".", handleRequest)

	tcp_server := &dns.Server{Addr: addr, Net: "tcp", Handler: handler}
	go func() {
		err := tcp_server.ListenAndServe()
		if err != nil {
			log.Fatal("Listen failed: ", err)
			os.Exit(1)
		}
	}()

	udp_server := &dns.Server{Addr: addr, Net: "udp", Handler: handler}
	log.Info("Start overture on " + addr)
	err := udp_server.ListenAndServe()
	if err != nil {
		log.Fatal("Listen failed: ", err)
		os.Exit(1)
	}
}

func handleRequest(w dns.ResponseWriter, q *dns.Msg) {

	inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	o := outbound.NewOutbound(q, inboundIP)
	m := config.Config.CachePool.Hit(q.Question[0], o.IPUsed, q.Id); if m != nil{
		w.WriteMsg(m)
		return
	}
	s := switcher.NewSwitcher(o)
	isAlternative := s.ChooseDNS()

	if err := o.ExchangeFromRemote(!isAlternative); err != nil {
		log.Debug("Get dns response failed: ", err)
		return
	}
	if !isAlternative {
		s.HandleResponseFromPrimaryDNS()
	} else {
		log.Debug("Finally use alternative DNS")
	}
	o.HandleMinimumTTL()
	common.LogAnswer(o.ResponseMessage)
	w.WriteMsg(o.ResponseMessage)
	config.Config.CachePool.InsertMessage(cache.Key(q.Question[0], o.IPUsed), o.ResponseMessage)
}
