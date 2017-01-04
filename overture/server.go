package overture

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"reflect"
)

func initServer() {

	handler := dns.NewServeMux()
	handler.HandleFunc(".", handleRequest)
	server := &dns.Server{Addr: Config.BindAddress, Net: "udp", Handler:handler}
	err := server.ListenAndServe()
	if err != nil{
		log.Fatal("Listen failed: ", err)
	}
}

func handleRequest(writer dns.ResponseWriter, question_message *dns.Msg) {

	temp_dns_server := chooseDNSServer(question_message)
	response_message := new(dns.Msg)
	err := getResponse(response_message, question_message, temp_dns_server)
	if err != nil {
		log.Warn("Get dns response failed: ", err)
		return
	}
	if reflect.DeepEqual(temp_dns_server, Config.PrimaryDNSServer) {
		MatchIPNetwork(response_message, question_message, ip_net_list)
	} else {
		log.Debug("Finally use alternative DNS.")
	}
	logAnswer(response_message)
	writer.WriteMsg(response_message)
}
