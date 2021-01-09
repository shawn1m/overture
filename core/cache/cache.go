// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package cache implements dns cache feature with edns-client-subnet support.
package cache

// Cache that holds RRs.

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// Elem hold an answer and additional section that returned from the cache.
// The signature is put in answer, extra is empty there. This wastes some memory.
type elem struct {
	expiration time.Time // time added + TTL, after this the elem is invalid
	msg        *dns.Msg
}

type elemData struct {
	Expiration time.Time
	Msg        []byte // dns.Msg cannot be converted to the json format successfully thus using its pack() method instead
}

func (e *elem) MarshalBinary() (data []byte, err error) {
	msgBytes, _ := e.msg.Pack()
	ed := elemData{e.expiration, msgBytes}
	return json.Marshal(ed)
}

func (e *elem) UnmarshalBinary(data []byte) error {
	var ed elemData
	err := json.Unmarshal(data, &ed)
	if err != nil {
		return err
	}
	e.expiration = ed.Expiration
	e.msg = &dns.Msg{}
	return e.msg.Unpack(ed.Msg)
}

// Cache is a cache that holds on the a number of RRs or DNS messages. The cache
// eviction is randomized.
type Cache struct {
	sync.RWMutex

	capacity    int
	table       map[string]*elem
	redisClient *redis.Client
}

// New returns a new cache with the capacity and the ttl specified.
func New(capacity int, redisUrl string, cacheRedisConnectionPoolSize int) *Cache {
	if capacity <= 0 {
		return nil
	}
	c := new(Cache)
	c.table = make(map[string]*elem)
	c.capacity = capacity

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		if redisUrl != "" {
			log.Error("redisUrl error ", redisUrl, err)
		}
	} else {
		if cacheRedisConnectionPoolSize > 0 {
			opt.PoolSize = cacheRedisConnectionPoolSize
		} else {
			log.Warn("cacheRedisConnectionPoolSize is ignored", cacheRedisConnectionPoolSize)
		}
		c.redisClient = redis.NewClient(opt)
		log.Info("Cache redis connected! ", c.redisClient.String())
	}

	return c
}

func (c *Cache) Capacity() int { return c.capacity }

func (c *Cache) Remove(s string) {
	if c.redisClient != nil {
		return
	}
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
	var err error
	if c.redisClient == nil {
		c.InsertMessageToLocal(s, m, mTTL)
	} else {
		err = c.InsertMessageToRedis(s, m, mTTL)
	}
	if err!=nil{
		log.Warnf("Insert cache failed: %s", s, err)
	}else {
		log.Debugf("Cached: %s", s)
	}
}

func (c *Cache) InsertMessageToRedis(s string, m *dns.Msg, mTTL uint32) error{

	ttlDuration := convertToTTLDuration(m, mTTL)
	if _, ok := c.table[s]; !ok {
		e := &elem{time.Now().Add(ttlDuration), m.Copy()}
		cmd := c.redisClient.Set(context.TODO(), s, e, ttlDuration)
		if cmd.Err() != nil {
			log.Warn("Redis set for cache failed!", cmd.Err())
			return cmd.Err()
		}
	}
	return nil

}
func (c *Cache) InsertMessageToLocal(s string, m *dns.Msg, mTTL uint32) {

	c.Lock()
	ttlDuration := convertToTTLDuration(m, mTTL)
	if _, ok := c.table[s]; !ok {
		e := &elem{time.Now().Add(ttlDuration), m.Copy()}
		c.table[s] = e
	}

	c.EvictRandom()
	c.Unlock()
}

func convertToTTLDuration(m *dns.Msg, mTTL uint32) time.Duration {
	var ttl uint32
	if len(m.Answer) == 0 {
		ttl = mTTL
	} else {
		ttl = m.Answer[0].Header().Ttl
	}
	return time.Duration(ttl) * time.Second
}

// Search returns a dns.Msg, the expiration time and a boolean indicating if we found something
// in the cache.
func (c *Cache) Search(s string) (*dns.Msg, time.Time, bool) {
	if c.capacity <= 0 {
		return nil, time.Time{}, false
	}
	if c.redisClient == nil {
		return c.SearchFromLocal(s)
	} else {
		return c.SearchFromRedis(s)
	}
}

func (c *Cache) SearchFromRedis(s string) (*dns.Msg, time.Time, bool) {
	var e elem
	err := c.redisClient.Get(context.TODO(), s).Scan(&e)
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Debug("Redis get return nil for ", s, err)
		} else {
			log.Warn("Redis get return nil for ", s, err)
		}
		return nil, time.Time{}, false
	}
	return e.msg, e.expiration, true

}

// todo: use finder implementation
func (c *Cache) SearchFromLocal(s string) (*dns.Msg, time.Time, bool) {
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
	return fmt.Sprintf("%s %d %s", q.Name, q.Qtype, ednsIP)
}

// Hit returns a dns message from the cache. If the message's TTL is expired, nil
// will be returned and the message is removed from the cache.
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

// Dump returns all local dns cache information for debugging
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
		var vs []string

		for _, a := range e.msg.Answer {
			vs = append(vs, a.String())
		}
		rs[k] = vs
	}
	return
}
