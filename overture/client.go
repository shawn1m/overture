package overture

import (
	log "github.com/Sirupsen/logrus"
	"net"
)

func getResponse(question_data []byte, address string) []byte {

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Warn("Can't resolve address: ", err)
		return nil
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Warn("Can't dial: ", err)
		return nil
	}

	defer conn.Close()
	conn.Write(question_data)

	response_buffer := make([]byte, 512)
	n, err := conn.Read(response_buffer)
	if err != nil {
		log.Warn("Read UDP message failed: ", err)
		return nil
	}

	return response_buffer[:n]
}
