/*
 * Copyright (c) 2019 shawn1m. All rights reserved.
 * Use of this source code is governed by The MIT License (MIT) that can be
 * found in the LICENSE file..
 */

// Package outbound implements multiple dns client and dispatcher for outbound connection.
package clients

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/shawn1m/overture/core/cache"
)

type CacheClient struct {
	responseMessage *dns.Msg
	questionMessage *dns.Msg

	ednsClientSubnetIP string

	cache *cache.Cache
}

func NewCacheClient(q *dns.Msg, ip string, cache *cache.Cache) *CacheClient {

	c := &CacheClient{questionMessage: q.Copy(), ednsClientSubnetIP: ip, cache: cache}
	return c
}

func (c *CacheClient) Exchange() *dns.Msg {

	if c.exchangeFromCache() {
		return c.responseMessage
	}

	return nil
}

func (c *CacheClient) exchangeFromCache() bool {

	if c.cache == nil {
		return false
	}

	m := c.cache.Hit(cache.Key(c.questionMessage.Question[0], c.ednsClientSubnetIP), c.questionMessage.Id)
	if m != nil {
		log.Debug("Cache Hit: " + cache.Key(c.questionMessage.Question[0], c.ednsClientSubnetIP))
		c.responseMessage = m
		return true
	}

	return false
}
