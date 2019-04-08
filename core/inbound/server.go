// Package inbound implements dns server for inbound connection.
package inbound

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/outbound"
)

type Server struct {
	bindAddress      string
	debugHttpAddress string
	dispatcher       outbound.Dispatcher
	rejectQType      []uint16
}

func NewServer(bindAddress string, debugHTTPAddress string, dispatcher outbound.Dispatcher, rejectQType []uint16) *Server {
	return &Server{
		bindAddress:      bindAddress,
		debugHttpAddress: debugHTTPAddress,
		dispatcher:       dispatcher,
		rejectQType:      rejectQType,
	}
}

func (s *Server) DumpCache(w http.ResponseWriter, req *http.Request) {
	if s.dispatcher.Cache == nil {
		io.WriteString(w, "error: cache not enabled")
		return
	}

	type answer struct {
		Name  string `json:"name"`
		TTL   int    `json:"ttl"`
		Type  string `json:"type"`
		Rdata string `json:"rdata"`
	}

	type response struct {
		Length   int                  `json:"length"`
		Capacity int                  `json:"capacity"`
		Body     map[string][]*answer `json:"body"`
	}

	query := req.URL.Query()
	nobody := true
	if t := query.Get("nobody"); strings.ToLower(t) == "false" {
		nobody = false
	}

	rs, l := s.dispatcher.Cache.Dump(nobody)
	body := make(map[string][]*answer)

	for k, es := range rs {
		answers := []*answer{}
		for _, e := range es {
			ts := strings.Split(e, "\t")
			ttl, _ := strconv.Atoi(ts[1])
			r := &answer{
				Name:  ts[0],
				TTL:   ttl,
				Type:  ts[3],
				Rdata: ts[4],
			}
			answers = append(answers, r)
		}
		body[strings.TrimSpace(k)] = answers
	}

	res := response{
		Body:     body,
		Length:   l,
		Capacity: s.dispatcher.Cache.Capacity(),
	}

	responseBytes, err := json.Marshal(&res)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	io.WriteString(w, string(responseBytes))
}

func (s *Server) Run() {

	mux := dns.NewServeMux()
	mux.Handle(".", s)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	log.Info("Start overture on " + s.bindAddress)

	for _, p := range [2]string{"tcp", "udp"} {
		go func(p string) {
			err := dns.ListenAndServe(s.bindAddress, p, mux)
			if err != nil {
				log.Fatal("Listen "+p+" failed: ", err)
				os.Exit(1)
			}
		}(p)
	}

	if s.debugHttpAddress != "" {
		http.HandleFunc("/cache", s.DumpCache)
		wg.Add(1)
		go http.ListenAndServe(s.debugHttpAddress, nil)
	}

	wg.Wait()
}

func (s *Server) ServeDNS(w dns.ResponseWriter, q *dns.Msg) {
	inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	log.Debug("Question from " + inboundIP + ": " + q.Question[0].String())

	for _, qt := range s.rejectQType {
		if isQuestionType(q, qt) {
			return
		}
	}

	responseMessage := s.dispatcher.Exchange(q, inboundIP)

	if responseMessage == nil {
		dns.HandleFailed(w, q)
		return
	}

	err := w.WriteMsg(responseMessage)
	if err != nil {
		log.Warn("Write message fail:", responseMessage)
		return
	}
}

func isQuestionType(q *dns.Msg, qt uint16) bool { return q.Question[0].Qtype == qt }
