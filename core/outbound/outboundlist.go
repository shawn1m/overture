// Copyright (c) 2016 holyshawn. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package outbound

import (
	"github.com/holyshawn/overture/core/config"
	"github.com/miekg/dns"
)

type OutboundListType struct {
	OutboundList []*Outbound

	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	DNSUpstream []*config.DNSUpstream
	InboundIP   string
}

func NewOutboundList(q *dns.Msg, dl []*config.DNSUpstream, inboundIP string) *OutboundListType {

	ol := new(OutboundListType)

	ol.QuestionMessage = q
	ol.InboundIP = inboundIP
	ol.DNSUpstream = dl

	for _, u := range dl {

		o := NewOutbound(q, u, inboundIP)
		ol.OutboundList = append(ol.OutboundList, o)
	}

	return ol
}

func (ol *OutboundListType) ExchangeFromRemote(IsCache bool, isLog bool) {

	ch := make(chan *dns.Msg, len(ol.OutboundList))

	for _, o := range ol.OutboundList {
		go func(o *Outbound, ch chan *dns.Msg) {
			o.ExchangeFromRemote(IsCache, isLog)
			ch <- o.ResponseMessage
		}(o, ch)
	}

	for i := 0; i < len(ol.OutboundList); i++ {
		if m := <-ch; m != nil {
			ol.ResponseMessage = m
			return
		}
	}
}

func (ol *OutboundListType) ExchangeFromLocal() bool {

	for _, o := range ol.OutboundList {
		if o.ExchangeFromLocal() {
			ol.ResponseMessage = o.ResponseMessage
			o.LogAnswer(true)
			return true
		}
	}
	return false
}

func (ol *OutboundListType) UpdateDNSUpstream(dl []*config.DNSUpstream) {

	for _, u := range dl {
		for _, o := range ol.OutboundList {
			o.DNSUpstream = u
		}
	}
}

func (ol *OutboundListType) EqualDNSUpstream(dl []*config.DNSUpstream) bool {

	for _, u := range dl {
		for _, o := range ol.OutboundList {
			if o.DNSUpstream != u {
				return false
			}
		}
	}

	return true
}
