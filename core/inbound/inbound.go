// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package inbound

import (
	"net"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/config"
	"github.com/shawn1m/overture/core/dispatcher"
	"github.com/shawn1m/overture/core/outbound"
)

func InitServer(addr string) {

	s := &server{addr: addr}
	s.Run()
}

type server struct {
	addr string
}

func (s *server) Run() {

	mux := dns.NewServeMux()
	mux.Handle(".", s)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	log.Info("Start overture on " + s.addr)

	for _, p := range [2]string{"tcp", "udp"} {
		go func(p string) {
			err := dns.ListenAndServe(s.addr, p, mux)
			if err != nil {
				log.Fatal("Listen "+p+" failed: ", err)
				os.Exit(1)
			}
		}(p)
	}

	wg.Wait()
}

func (s *server) ServeDNS(w dns.ResponseWriter, q *dns.Msg) {

	inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	ob := outbound.NewBundle(q, config.Config.PrimaryDNS, inboundIP)

	log.Debug("Question: " + ob.QuestionMessage.Question[0].String())

	for _, qt := range config.Config.RejectQtype {
		if isQuestionType(q, qt) {
			return
		}
	}

	if ok := ob.ExchangeFromLocal(); ok {
		if ob.ResponseMessage != nil {
			w.WriteMsg(ob.ResponseMessage)
			return
		}
	}

	func() {
		if config.Config.OnlyPrimaryDNS {
			ob.ExchangeFromRemote(true, true)
		} else {
			d := dispatcher.New(ob)
			if d.ExchangeForIPv6() || d.ExchangeForDomain() {
				return
			}

			ob.ExchangeFromRemote(false, true)
			d.ExchangeForPrimaryDNSResponse()
		}
	}()

	if ob.ResponseMessage != nil {
		w.WriteMsg(ob.ResponseMessage)
	}
}

func isQuestionType(q *dns.Msg, qt uint16) bool { return q.Question[0].Qtype == qt }
