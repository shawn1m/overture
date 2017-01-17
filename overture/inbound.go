package overture

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"net"
	"os"
	"reflect"
)

func initServer() {
	handler := dns.NewServeMux()
	handler.HandleFunc(".", handleRequest)

	tcp_server := &dns.Server{Addr: Config.BindAddress, Net: "tcp", Handler: handler}
	go func() {
		err := tcp_server.ListenAndServe()
		if err != nil {
			log.Fatal("Listen failed: ", err)
			os.Exit(1)
		}
	}()

	udp_server := &dns.Server{Addr: Config.BindAddress, Net: "udp", Handler: handler}
	log.Info("Start overture on " + Config.BindAddress)
	err := udp_server.ListenAndServe()
	if err != nil {
		log.Fatal("Listen failed: ", err)
		os.Exit(1)
	}
}

func handleRequest(writer dns.ResponseWriter, question_message *dns.Msg) {

	temp_dns_server := new(dnsServer)
	DNSServerFilter(temp_dns_server, question_message)
	response_message := new(dns.Msg)
	remote_ip, _, _ := net.SplitHostPort(writer.RemoteAddr().String())
	err := getResponse(response_message, question_message, remote_ip, *temp_dns_server)
	if err != nil {
		if err == dns.ErrTruncated {
			log.Warn("Maybe your primary dns server does not support edns client subnet: ", err)
		} else {
			log.Warn("Get dns response failed: ", err)
		}
		return
	}
	if reflect.DeepEqual(temp_dns_server, Config.PrimaryDNSServer) {
		PrimaryDNSResponseFilter(response_message, question_message, remote_ip, Config.IPNetworkList)
	} else {
		log.Debug("Finally use alternative DNS")
	}
	MinimumTTLFilter(response_message, uint32(Config.MinimumTTL))
	logAnswer(response_message)
	writer.WriteMsg(response_message)
}
