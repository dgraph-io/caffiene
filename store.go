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

import "sync"

// store is the interface fulfilled by all hash map implementations in this
// file. Some hash map implementations are better suited for certain data
// distributions than others, so this allows us to abstract that out for use
// in Ristretto.
//
// Every store is safe for concurrent usage.
type store interface {
	// Get returns the value associated with the key parameter.
	Get(uint64) interface{}
	// Set adds the key-value pair to the Map or updates the value if it's
	// already present.
	Set(uint64, interface{})
	// Del deletes the key-value pair from the Map.
	Del(uint64)
	// Run applies the function parameter to random key-value pairs. No key
	// will be visited more than once. If the function returns false, the
	// iteration stops. If the function returns true, the iteration will
	// continue until every key has been visited once.
	Run(func(interface{}, interface{}) bool)
}

// newStore returns the default store implementation.
func newStore() store {
	return newSyncMap()
}

type syncMap struct {
	*sync.Map
}

func newSyncMap() store {
	return &syncMap{&sync.Map{}}
}

func (m *syncMap) Get(key uint64) interface{} {
	value, _ := m.Load(key)
	return value
}

func (m *syncMap) Set(key uint64, value interface{}) {
	m.Store(key, value)
}

func (m *syncMap) Del(key uint64) {
	m.Delete(key)
}

func (m *syncMap) Run(f func(key, value interface{}) bool) {
	m.Range(f)
}

type lockedMap struct {
	sync.RWMutex
	data map[uint64]interface{}
}

func newLockedMap() *lockedMap {
	return &lockedMap{data: make(map[uint64]interface{})}
}

func (m *lockedMap) Get(key uint64) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.data[key]
}

func (m *lockedMap) Set(key uint64, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
}

func (m *lockedMap) Del(key uint64) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
}

func (m *lockedMap) Run(f func(interface{}, interface{}) bool) {
	m.RLock()
	defer m.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			return
		}
	}
}
