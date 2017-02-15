package inbound

import (
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/outbound"
	"github.com/holyshawn/overture/core/switcher"
	"github.com/miekg/dns"
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
	ol := outbound.NewOutboundList(q, config.Config.PrimaryDNS, inboundIP)

	log.Debug("Question: " + ol.QuestionMessage.Question[0].String())

	if ol.ExchangeFromLocal(){
		if ol.ResponseMessage != nil {
			w.WriteMsg(ol.ResponseMessage)
			return
		}
	}

	s := switcher.NewSwitcher(ol)

	func() {
		if s.ChooseDNS() {
			return
		}

		s.HandleResponseFromPrimaryDNS()
	}()

	if ol.ResponseMessage != nil {
		w.WriteMsg(ol.ResponseMessage)
	}
}
