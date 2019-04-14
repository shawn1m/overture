// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package cache implements dns cache feature with edns-client-subnet support.
package cache

// Cache that holds RRs.

import (
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// Elem hold an answer and additional section that returned from the cache.
// The signature is put in answer, extra is empty there. This wastes some memory.
type elem struct {
	expiration time.Time // time added + TTL, after this the elem is invalid
	msg        *dns.Msg
}

// Cache is a cache that holds on the a number of RRs or DNS messages. The cache
// eviction is randomized.
type Cache struct {
	sync.RWMutex

	capacity int
	table    map[string]*elem
}

// New returns a new cache with the capacity and the ttl specified.
func New(capacity int) *Cache {
	if capacity <= 0 {
		return nil
	}
	c := new(Cache)
	c.table = make(map[string]*elem)
	c.capacity = capacity
	return c
}

func (c *Cache) Capacity() int { return c.capacity }

func (c *Cache) Remove(s string) {
	c.Lock()
	delete(c.table, s)
	c.Unlock()
}

// EvictRandom removes a random member a the cache.
// Must be called under a write lock.
func (c *Cache) EvictRandom() {
	cacheLength := len(c.table)
	if cacheLength <= c.capacity {
		return
	}
	i := c.capacity - cacheLength
	for k := range c.table {
		delete(c.table, k)
		i--
		if i == 0 {
			break
		}
	}
}

// InsertMessage inserts a message in the Cache. We will cache it for ttl seconds, which
// should be a small (60...300) integer.
func (c *Cache) InsertMessage(s string, m *dns.Msg, mTTL uint32) {
	if c.capacity <= 0 || m == nil {
		return
	}

	c.Lock()
	var ttl uint32
	if len(m.Answer) == 0 {
		ttl = mTTL
	} else {
		ttl = m.Answer[0].Header().Ttl
	}
	ttlDuration := time.Duration(ttl) * time.Second
	if _, ok := c.table[s]; !ok {
		c.table[s] = &elem{time.Now().UTC().Add(ttlDuration), m.Copy()}
	}
	log.Debug("Cached: " + s)
	c.EvictRandom()
	c.Unlock()
}

// Search returns a dns.Msg, the expiration time and a boolean indicating if we found something
// in the cache.
// todo: use finder implementation
func (c *Cache) Search(s string) (*dns.Msg, time.Time, bool) {
	if c.capacity <= 0 {
		return nil, time.Time{}, false
	}
	c.RLock()
	if e, ok := c.table[s]; ok {
		e1 := e.msg.Copy()
		c.RUnlock()
		return e1, e.expiration, true
	}
	c.RUnlock()
	return nil, time.Time{}, false
}

// Key creates a hash key from a question section.
func Key(q dns.Question, ednsIP string) string {
	return q.Name + " " + strconv.Itoa(int(q.Qtype)) + " " + ednsIP
}

// Hit returns a dns message from the cache. If the message's TTL is expired nil
// is returned and the message is removed from the cache.
func (c *Cache) Hit(key string, msgid uint16) *dns.Msg {
	m, exp, hit := c.Search(key)
	if hit {
		// Cache hit! \o/
		if time.Since(exp) < 0 {
			m.Id = msgid
			m.Compress = true
			// Even if something ended up with the TC bit *in* the cache, set it to off
			m.Truncated = false
			for _, a := range m.Answer {
				a.Header().Ttl = uint32(time.Since(exp).Seconds() * -1)
			}
			return m
		}
		// Expired! /o\
		c.Remove(key)
	}
	return nil
}

// Dump returns all dns cache information, for dubugging
func (c *Cache) Dump(nobody bool) (rs map[string][]string, l int) {
	if c.capacity <= 0 {
		return
	}

	l = len(c.table)

	rs = make(map[string][]string)

	if nobody {
		return
	}

	c.RLock()
	defer c.RUnlock()

	for k, e := range c.table {
		vs := []string{}

		for _, a := range e.msg.Answer {
			vs = append(vs, a.String())
		}
		rs[k] = vs
	}
	return
}
