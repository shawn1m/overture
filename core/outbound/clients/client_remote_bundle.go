/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

package clients

import (
	"github.com/miekg/dns"
	"github.com/shawn1m/overture/core/cache"
	"github.com/shawn1m/overture/core/common"
)

type RemoteClientBundle struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	clients []*RemoteClient

	dnsUpstreams []*common.DNSUpstream
	inboundIP    string
	minimumTTL   int

	cache *cache.Cache
}

func NewClientBundle(q *dns.Msg, ul []*common.DNSUpstream, ip string, minimumTTL int, cache *cache.Cache) *RemoteClientBundle {

	cb := &RemoteClientBundle{questionMessage: q.Copy(), dnsUpstreams: ul, inboundIP: ip, minimumTTL: minimumTTL, cache: cache}

	for _, u := range ul {

		c := NewClient(cb.questionMessage, u, cb.inboundIP, cb.cache)
		cb.clients = append(cb.clients, c)
	}

	return cb
}

func (cb *RemoteClientBundle) Exchange(isCache bool, isLog bool) *dns.Msg {

	ch := make(chan *RemoteClient, len(cb.clients))

	for _, o := range cb.clients {
		go func(c *RemoteClient, ch chan *RemoteClient) {
			c.Exchange(isLog)
			ch <- c
		}(o, ch)
	}

	var ec *RemoteClient

	for i := 0; i < len(cb.clients); i++ {
		if c := <-ch; c.responseMessage != nil {
			ec = c
			if common.HasAnswer(c.responseMessage) {
				break
			}
		}
	}

	if ec != nil && ec.responseMessage != nil {
		cb.responseMessage = ec.responseMessage
		cb.questionMessage = ec.questionMessage

		if isCache {
			cb.CacheResult()
		}
	}

	return cb.responseMessage
}

func (cb *RemoteClientBundle) CacheResult() {

	if cb.cache != nil {
		cb.cache.InsertMessage(cache.Key(cb.questionMessage.Question[0], common.GetEDNSClientSubnetIP(cb.questionMessage)), cb.responseMessage)
	}
}

func (cb *RemoteClientBundle) setMinimumTTL() {

	minimumTTL := uint32(cb.minimumTTL)
	if minimumTTL == 0 {
		return
	}
	for _, a := range cb.responseMessage.Answer {
		if a.Header().Ttl < minimumTTL {
			a.Header().Ttl = minimumTTL
		}
	}
}

func (cb *RemoteClientBundle) IsType(t uint16) bool {
	return t == cb.questionMessage.Question[0].Qtype
}

func (cb *RemoteClientBundle) GetFirstQuestionDomain() string {
	return cb.questionMessage.Question[0].Name[:len(cb.questionMessage.Question[0].Name)-1]
}

func (cb *RemoteClientBundle) GetResponseMessage() *dns.Msg {
	return cb.responseMessage
}
