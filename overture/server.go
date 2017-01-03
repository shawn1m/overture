package overture

import (
	log "github.com/Sirupsen/logrus"
	"net"
	"github.com/miekg/dns"
)


func ListenAndReceive(address string, nbWorkers int) error {

       c, err := net.ListenPacket("udp", address)
       if err != nil {
            return err
       }

       for i := 0; i < nbWorkers; i++ {
           go func() {
               handleQuestion(c)
            }()
       }
       return nil
}

func handleQuestion(conn net.PacketConn) {

	question_buffer := make([]byte, 512)
	n, addr, err := conn.ReadFrom(question_buffer)
	if err != nil {
		log.Warn("Read question message failed. ", err)
		return
	}

	go func() {
		question_message := new(dns.Msg)
		question_message.Unpack(question_buffer[:n])
		temp_dns_addr := chooseDNSAddr(question_message)
		response_message := getResponse("tcp", question_message, temp_dns_addr)
		if temp_dns_addr == Config.PrimaryDNSAddress {
			ResponseMatchIPNetwork(response_message, question_message, ip_net_list)
		} else {
			log.Debug("Finally use alternative DNS.")
		}
		logResponse(response_message)
		b, _ := response_message.Pack()
		conn.WriteTo(b, addr)
	}()
}
