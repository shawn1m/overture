// Copyright (c) 2016 holyshawn. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

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
	ob := outbound.NewOutboundBundle(q, config.Config.PrimaryDNS, inboundIP)

	log.Debug("Question: " + ob.QuestionMessage.Question[0].String())

	if ob.ExchangeFromLocal() {
		if ob.ResponseMessage != nil {
			w.WriteMsg(ob.ResponseMessage)
			return
		}
	}

	func() {
		s := switcher.NewSwitcher(ob)
		if s.ExchangeForIPv6() || s.ExchangeForDomain() {
			return
		}

		ob.ExchangeFromRemote(false, true)
		s.ExchangeForPrimaryDNSResponse()
	}()

	if ob.ResponseMessage != nil {
		w.WriteMsg(ob.ResponseMessage)
	}
}
