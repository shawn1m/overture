// Package inbound implements dns server for inbound connection.
package inbound

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	"github.com/coredns/coredns/plugin/pkg/doh"
	"github.com/coredns/coredns/plugin/pkg/response"
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/common"
	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/outbound"
)

type Server struct {
	bindAddress      string
	debugHttpAddress string
	dispatcher       outbound.Dispatcher
	rejectQType      []uint16
	HTTPMux          *http.ServeMux
	ctx              context.Context
	cancel           context.CancelFunc
	dohEnabled       bool
}

func NewServer(bindAddress string, debugHTTPAddress string, dispatcher outbound.Dispatcher, rejectQType []uint16, dohEnabled bool) *Server {
	s := &Server{
		bindAddress:      bindAddress,
		debugHttpAddress: debugHTTPAddress,
		dispatcher:       dispatcher,
		rejectQType:      rejectQType,
		dohEnabled:       dohEnabled,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.HTTPMux = http.NewServeMux()
	return s
}

func (s *Server) ServeDNSHttp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != doh.Path {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	q, err := doh.RequestToMsg(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a DoHWriter with the correct addresses in it.
	inboundIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	forwardIP := r.Header.Get("X-Forwarded-For")
	if net.ParseIP(forwardIP) != nil && common.ReservedIPNetworkList.Contains(net.ParseIP(inboundIP), false, "") {
		inboundIP = forwardIP
	}
	log.Debugf("Question from %s: %s", inboundIP, q.Question[0].String())

	for _, qt := range s.rejectQType {
		if isQuestionType(q, qt) {
			log.Debugf("Reject %s: %s", inboundIP, q.Question[0].String())
			http.Error(w, "Rejected", http.StatusForbidden)
			return
		}
	}

	responseMessage := s.dispatcher.Exchange(q, inboundIP)

	if responseMessage == nil {
		http.Error(w, "No response", http.StatusInternalServerError)
		return
	}

	buf, _ := responseMessage.Pack()

	mt, _ := response.Typify(responseMessage, time.Now().UTC())
	age := dnsutil.MinimalTTL(responseMessage, mt)

	w.Header().Set("Content-Type", doh.MimeType)
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%f", age.Seconds()))
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.WriteHeader(http.StatusOK)

	w.Write(buf)
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
		var answers []*answer
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

	log.Infof("Overture is listening on %s", s.bindAddress)

	for _, p := range [2]string{"tcp", "udp"} {
		go func(p string) {

			// Manual create server inorder to have a way to close it.
			srv := &dns.Server{Addr: s.bindAddress, Net: p, Handler: mux}
			go func() {
				<-s.ctx.Done()
				log.Warnf("Shutting down the server on protocol %s", p)
				srv.ShutdownContext(s.ctx)
			}()
			err := srv.ListenAndServe()
			if err != nil {
				log.Fatalf("Listening on port %s failed: %s", p, err)
				os.Exit(1)
			}
			wg.Done()
		}(p)
	}

	if s.debugHttpAddress != "" {
		s.HTTPMux.HandleFunc("/cache", s.DumpCache)
		s.HTTPMux.HandleFunc("/debug/pprof/", pprof.Index)
		s.HTTPMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		s.HTTPMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		s.HTTPMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		s.HTTPMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		if s.dohEnabled {
			log.Info("Dns over http server started!")
			s.HTTPMux.HandleFunc(doh.Path, s.ServeDNSHttp)
		}

		wg.Add(1)
		go func() {
			// Manual create server inorder to have a way to close it.
			srv := &http.Server{
				Addr:    s.debugHttpAddress,
				Handler: s.HTTPMux,
			}
			go func() {
				<-s.ctx.Done()
				log.Warnf("Shutting down debug HTTP server")
				srv.Shutdown(s.ctx)
			}()

			err := srv.ListenAndServe()
			if err != http.ErrServerClosed {
				log.Fatalf("Debug HTTP Server Listen on port %s  faild: %s", s.debugHttpAddress, err)
				os.Exit(1)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func (s *Server) Stop() {
	s.cancel()
}

func (s *Server) ServeDNS(w dns.ResponseWriter, q *dns.Msg) {
	inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	log.Debugf("Question from %s: %s", inboundIP, q.Question[0].String())

	for _, qt := range s.rejectQType {
		if isQuestionType(q, qt) {
			log.Debugf("Reject %s: %s", inboundIP, q.Question[0].String())
			dns.HandleFailed(w, q)
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
		log.Warnf("Write message failed, message: %s, error: %s", responseMessage, err)
		return
	}
}

func isQuestionType(q *dns.Msg, qt uint16) bool { return q.Question[0].Qtype == qt }
