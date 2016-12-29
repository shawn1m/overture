package overture

import (
	log "github.com/Sirupsen/logrus"
	"net"
)

func handleUDPQuestion(conn *net.UDPConn) {

	question_buffer := make([]byte, 512)
	n, remote_addr, err := conn.ReadFromUDP(question_buffer)
	if err != nil {
		log.Warn("Read UDP message failed. ", err)
		return
	}

	go func() {
		temp_dns_addr := chooseDNSAddr(question_buffer[:n])
		response_data := getResponse(question_buffer[:n], temp_dns_addr)
		if temp_dns_addr == Config.PrimaryDNSAddress {
			MatchDomesticIPResponse(&response_data, question_buffer[:n])
		} else {
			log.Debug("Finally use alternative DNS.")
		}
		logResponseAnswer(response_data)
		conn.WriteToUDP(response_data, remote_addr)
	}()
}
