// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package outbound

import (
	"github.com/shawn1m/overture/core/config"
	"github.com/miekg/dns"
)

type OutboundBundle struct {
	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	bundleList   []*outbound
	upstreamList []*config.DNSUpstream
	inboundIP    string
}

func NewOutboundBundle(q *dns.Msg, ul []*config.DNSUpstream, inboundIP string) *OutboundBundle {

	ob := new(OutboundBundle)

	ob.QuestionMessage = q
	ob.inboundIP = inboundIP
	ob.upstreamList = ul

	for _, u := range ul {

		o := newOutbound(q, u, inboundIP)
		ob.bundleList = append(ob.bundleList, o)
	}

	return ob
}

func (ob *OutboundBundle) ExchangeFromRemote(isCache bool, isLog bool) {

	ch := make(chan *dns.Msg, len(ob.bundleList))

	for _, o := range ob.bundleList {
		go func(o *outbound, ch chan *dns.Msg) {
			o.exchangeFromRemote(isCache, isLog)
			ch <- o.ResponseMessage
		}(o, ch)
	}

	for i := 0; i < len(ob.bundleList); i++ {
		if m := <-ch; m != nil {
			ob.ResponseMessage = m
			return
		}
	}
}

func (ob *OutboundBundle) ExchangeFromLocal() bool {

	for _, o := range ob.bundleList {
		if o.exchangeFromLocal() {
			ob.ResponseMessage = o.ResponseMessage
			o.logAnswer(true)
			return true
		}
	}
	return false
}

func (ob *OutboundBundle) UpdateFromDNSUpstream(ul []*config.DNSUpstream) {

	ob.upstreamList = ul
	ob.ResponseMessage = nil

	var ol []*outbound

	for _, u := range ul {
		o := newOutbound(ob.QuestionMessage, u, ob.inboundIP)
		o.QuestionMessage = ob.QuestionMessage
		ol = append(ol, o)
	}

	ob.bundleList = ol
}

func (ob *OutboundBundle) EqualDNSUpstream(ul []*config.DNSUpstream) bool {

	for _, u := range ul {
		for _, o := range ob.bundleList {
			if o.DNSUpstream != u {
				return false
			}
		}
	}

	return true
}
