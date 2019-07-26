/*
 * Copyright 2019 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ristretto

import "github.com/dgraph-io/ristretto/z"

// Cache ties everything together. The three main components are:
//
//     1) The hash map: this is the Map interface.
//     2) The admission and eviction policy: this is the Policy interface.
//     3) The bp-wrapper buffer: this is the Buffer struct.
//
// All three of these components work together to try and keep the most valuable
// key-value pairs in the hash map. Value is determined by the Policy, and
// BP-Wrapper keeps the Policy fast (by batching metadata updates).
type Cache struct {
	data   store
	policy Policy
	buffer *ringBuffer
	notify func(uint64)
}

type Config struct {
	// NumCounters is the number of counters to preallocate space for. This
	// should correspond to the expected number of items in the Cache when it
	// is full (when MaxCost is reached).
	NumCounters int64
	// MaxCost is the cost capacity of the Cache.
	MaxCost int64
	// BufferItems is max number of items in access batches (BP-Wrapper).
	BufferItems int64
	// OnEvict is ran for each key evicted. We do not store key strings, so the
	// param is the AESHash of the key.
	OnEvict func(uint64)
	// Log is whether or not to Log hit ratio statistics (with some overhead).
	Log bool
}

func NewCache(config *Config) *Cache {
	if config.MaxCost == 0 && config.NumCounters != 0 {
		config.MaxCost = config.NumCounters
	}
	// data is the hash map for the entire cache, it's initialized outside of
	// the cache struct declaration because it may need to be passed to the
	// policy in some cases
	data := newStore()
	// initialize the policy (with a recorder wrapping if logging is enabled)
	policy := newPolicy(config.NumCounters, config.MaxCost)
	if config.Log {
		policy = NewRecorder(policy, data)
	}
	return &Cache{
		data:   data,
		policy: policy,
		buffer: newRingBuffer(ringLossy, &ringConfig{
			Consumer: policy,
			Capacity: config.BufferItems,
		}),
		notify: config.OnEvict,
	}
}

func (c *Cache) Get(key string) interface{} {
	hashed := z.AESHashString(key)
	c.buffer.Push(hashed)
	return c.data.Get(hashed)
}

func (c *Cache) Set(key string, val interface{}, cost int64) bool {
	hashed := z.AESHashString(key)
	victims, added := c.policy.Add(hashed, cost)
	if !added {
		return false
	}
	for _, victim := range victims {
		c.data.Del(victim)
		if c.notify != nil {
			c.notify(victim)
		}
	}
	c.data.Set(hashed, val)
	return true
}

func (c *Cache) Del(key string) {
	hashed := z.AESHashString(key)
	c.policy.Del(hashed)
	c.data.Del(hashed)
}

func (c *Cache) Log() *PolicyLog {
	return c.policy.Log()
}
