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
	domainTTLMap map[string]uint32

	cache *cache.Cache
	Name  string
}

func NewClientBundle(q *dns.Msg, ul []*common.DNSUpstream, ip string, minimumTTL int, cache *cache.Cache, name string, domainTTLMap map[string]uint32) *RemoteClientBundle {

	cb := &RemoteClientBundle{questionMessage: q.Copy(), dnsUpstreams: ul, inboundIP: ip, minimumTTL: minimumTTL, cache: cache, Name: name, domainTTLMap: domainTTLMap}

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
		c := <-ch
		if c != nil {
			ec = c
			break
		}
	}

	if ec != nil && ec.responseMessage != nil {
		cb.responseMessage = ec.responseMessage
		cb.questionMessage = ec.questionMessage

		common.SetMinimumTTL(cb.responseMessage, uint32(cb.minimumTTL))
		common.SetTTLByMap(cb.responseMessage, cb.domainTTLMap)

		if isCache {
			cb.CacheResultIfNeeded()
		}
	}

	return cb.responseMessage
}

func (cb *RemoteClientBundle) ExchangeFromCache() *dns.Msg {
	for _, o := range cb.clients {
		cb.responseMessage = o.ExchangeFromCache()
		if cb.responseMessage != nil {
			return cb.responseMessage
		}
	}
	return cb.responseMessage
}

func (cb *RemoteClientBundle) CacheResultIfNeeded() {

	if cb.cache != nil {
		cb.cache.InsertMessage(cache.Key(cb.questionMessage.Question[0], common.GetEDNSClientSubnetIP(cb.questionMessage)), cb.responseMessage, uint32(cb.minimumTTL))
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
