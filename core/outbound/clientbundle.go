// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package outbound

import (
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
	"github.com/shawn1m/overture/core/hosts"
)

type ClientBundle struct {
	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	ClientList []*Client

	DNSUpstreamList []*DNSUpstream
	InboundIP       string

	Hosts     *hosts.Hosts
	CachePool *cache.Cache
}

func NewClientBundle(q *dns.Msg, ul []*DNSUpstream, ip string, h *hosts.Hosts, cp *cache.Cache) *ClientBundle {

	cb := &ClientBundle{QuestionMessage: q, DNSUpstreamList: ul, InboundIP: ip, Hosts: h, CachePool: cp}

	for _, u := range ul {

		c := NewClient(cb.QuestionMessage, u, cb.InboundIP, cb.Hosts, cb.CachePool)
		cb.ClientList = append(cb.ClientList, c)
	}

	return cb
}

func (cb *ClientBundle) ExchangeFromRemote(isCache bool, isLog bool) {

	ch := make(chan *dns.Msg, len(cb.ClientList))

	for _, o := range cb.ClientList {
		go func(c *Client, ch chan *dns.Msg) {
			c.ExchangeFromRemote(isCache, isLog)
			ch <- c.ResponseMessage
		}(o, ch)
	}

	var em *dns.Msg

	for i := 0; i < len(cb.ClientList); i++ {
		if m := <-ch; m != nil {
			if common.IsAnswerEmpty(m) {
				em = m
				break
			}
			cb.ResponseMessage = m
			return
		}
	}
	cb.ResponseMessage = em
}

func (cb *ClientBundle) ExchangeFromLocal() bool {

	for _, c := range cb.ClientList {
		if c.ExchangeFromLocal() {
			cb.ResponseMessage = c.ResponseMessage
			c.logAnswer(true)
			return true
		}
	}
	return false
}

func (cb *ClientBundle) EqualDNSUpstream(ul []*DNSUpstream) bool {

	for _, u := range ul {
		for _, c := range cb.ClientList {
			if c.DNSUpstream != u {
				return false
			}
		}
	}

	return true
}

func (cb *ClientBundle) UpdateFromDNSUpstream(ul []*DNSUpstream) {

	cb.DNSUpstreamList = ul
	cb.ResponseMessage = nil

	var cl []*Client

	for _, u := range ul {
		c := NewClient(cb.QuestionMessage, u, cb.InboundIP, cb.Hosts, cb.CachePool)
		cl = append(cl, c)
	}

	cb.ClientList = cl
}
