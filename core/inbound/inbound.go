package inbound

import (
	"net"
	"os"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/switcher"
	"github.com/holyshawn/overture/core/outbound"
	"github.com/holyshawn/overture/core/common"
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
	o := outbound.NewOutbound(q, inboundIP, new(config.DNSUpstream))
	s := switcher.NewSwitcher(o)
	s.ChooseNameSever()

		if err := o.ExchangeFromRemote(); err != nil {
			log.Debug("Get dns response failed: ", err)
			return
		}
		if reflect.DeepEqual(o.DNSUpstream, config.Config.PrimaryDNSServer) {
			s.HandleResponseFromPrimaryDNS()
		} else {
			log.Debug("Finally use alternative DNS")
		}
		o.HandleMinimumTTL()
		common.LogAnswer(o.ResponseMessage)
		w.WriteMsg(o.ResponseMessage)
	}
