// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package cache

// Cache that holds RRs and for DNSSEC an RRSIG.

// TODO(miek): there is a lot of copying going on to copy myself out of data
// races. This should be optimized.

import (
	"crypto/sha1"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

// Elem hold an answer and additional section that returned from the cache.
// The signature is put in answer, extra is empty there. This wastes some memory.
type elem struct {
	expiration time.Time // time added + TTL, after this the elem is invalid
	m          *dns.Msg
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
	clen := len(c.table)
	if clen < c.capacity {
		return
	}
	i := c.capacity - clen
	for k, _ := range c.table {
		delete(c.table, k)
		i--
		if i == 0 {
			break
		}
	}
}

// InsertMessage inserts a message in the Cache. We will cache it for ttl seconds, which
// should be a small (60...300) integer.
func (c *Cache) InsertMessage(s string, m *dns.Msg) {
	if c.capacity <= 0 || len(m.Answer) == 0 {
		return
	}

	c.Lock()
	ttl := time.Duration(m.Answer[0].Header().Ttl) * time.Second
	if _, ok := c.table[s]; !ok {
		c.table[s] = &elem{time.Now().UTC().Add(ttl), m.Copy()}
	}
	log.Debug("Cache: " + s)
	c.EvictRandom()
	c.Unlock()
}

// Search returns a dns.Msg, the expiration time and a boolean indicating if we found something
// in the cache.
func (c *Cache) Search(s string) (*dns.Msg, time.Time, bool) {
	if c.capacity <= 0 {
		return nil, time.Time{}, false
	}
	c.RLock()
	if e, ok := c.table[s]; ok {
		e1 := e.m.Copy()
		c.RUnlock()
		return e1, e.expiration, true
	}
	c.RUnlock()
	return nil, time.Time{}, false
}

// Key creates a hash key from a question section. It creates a different key
// for requests with DNSSEC.
func Key(q dns.Question, ednsIP string) string {
	h := sha1.New()
	i := append([]byte(q.Name), packUint16(q.Qtype)...)
	i = append(i, []byte(ednsIP)...)
	return string(h.Sum(i))
}

func packUint16(i uint16) []byte { return []byte{byte(i >> 8), byte(i)} }
