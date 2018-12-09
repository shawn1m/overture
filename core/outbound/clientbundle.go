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

	DNSUpstreamList []*common.DNSUpstream
	InboundIP       string
	MinimumTTL      int

	Hosts *hosts.Hosts
	Cache *cache.Cache
}

func NewClientBundle(q *dns.Msg, ul []*common.DNSUpstream, ip string, ttl int, h *hosts.Hosts, cache *cache.Cache) *ClientBundle {

	cb := &ClientBundle{QuestionMessage: q.Copy(), DNSUpstreamList: ul, InboundIP: ip, MinimumTTL: ttl, Hosts: h, Cache: cache}

	for _, u := range ul {

		c := NewClient(cb.QuestionMessage, u, cb.InboundIP, cb.Hosts, cb.Cache)
		cb.ClientList = append(cb.ClientList, c)
	}

	return cb
}

func (cb *ClientBundle) ExchangeFromRemote(isCache bool, isLog bool) {

	ch := make(chan *Client, len(cb.ClientList))

	for _, o := range cb.ClientList {
		go func(c *Client, ch chan *Client) {
			c.ExchangeFromRemote(false, isLog)
			ch <- c
		}(o, ch)
	}

	var ec *Client

	for i := 0; i < len(cb.ClientList); i++ {
		if c := <-ch; c.ResponseMessage != nil {
			ec = c
			if common.HasAnswer(c.ResponseMessage) {
				break
			}
		}
	}

	if ec != nil && ec.ResponseMessage != nil {
		cb.ResponseMessage = ec.ResponseMessage
		cb.QuestionMessage = ec.QuestionMessage

		cb.setMinimumTTL()

		if isCache {
			cb.CacheResult()
		}
	}
}

func (cb *ClientBundle) ExchangeFromLocal() bool {

	for _, c := range cb.ClientList {
		if c.ExchangeFromLocal() {
			cb.ResponseMessage = c.ResponseMessage
			c.logAnswer("Local")
			return true
		}
	}
	return false
}

func (cb *ClientBundle) CacheResult() {

	if cb.Cache != nil {
		cb.Cache.InsertMessage(cache.Key(cb.QuestionMessage.Question[0], common.GetEDNSClientSubnetIP(cb.QuestionMessage)), cb.ResponseMessage)
	}
}

func (cb *ClientBundle) setMinimumTTL() {

	minimumTTL := uint32(cb.MinimumTTL)
	if minimumTTL == 0 {
		return
	}
	for _, a := range cb.ResponseMessage.Answer {
		if a.Header().Ttl < minimumTTL {
			a.Header().Ttl = minimumTTL
		}
	}
}
