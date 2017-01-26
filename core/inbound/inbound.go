package inbound

import (
	"github.com/holyshawn/overture/core/config"
	"github.com/holyshawn/overture/core/switcher"
	"github.com/holyshawn/overture/core/outbound"
	"github.com/holyshawn/overture/core/common"
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"net"
	"os"
	"reflect"
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

func handleRequest(writer dns.ResponseWriter, question_message *dns.Msg) {

	remote_ip, _, _ := net.SplitHostPort(writer.RemoteAddr().String())
	o := outbound.NewOutbound(question_message, remote_ip, new(config.DNSUpstream))
	s := switcher.NewSwitcher(o)
	s.ChooseNameSever()

		if err := o.ExchangeFromRemote(); err != nil {
			log.Debug("Get dns response failed: ", err)
			return
		}
		if reflect.DeepEqual(o.DomainNameServer, config.Config.PrimaryDNSServer) {
			s.HandleResponseFromPrimaryDNS()
		} else {
			log.Debug("Finally use alternative DNS")
		}
		o.HandleMinimumTTL()
		common.LogAnswer(o.ResponseMessage)
		writer.WriteMsg(o.ResponseMessage)
	}
