// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctex

import (
	"sync"
)

// Store is goroutine-safe key-value storage.
type Store = *store

// NewStore creates an empty [Store].
func NewStore() Store {
	return &store{}
}

type store struct {
	storage sync.Map
}

// Load returns the value stored under key or nil if key is absent.
func (s Store) Load(key any) any {
	v, ok := s.storage.Load(key)
	if !ok {
		return nil
	}
	if c, ok := v.(*cell); ok {
		v, ok := c.Read()
		if !ok {
			return nil
		}
		return v
	}
	return v
}

// Save stores value under key and returns the previous value.
func (s Store) Save(key any, value any) any {
	v, ok := s.storage.Swap(key, value)
	if !ok {
		return nil
	}
	if c, ok := v.(*cell); ok {
		v, ok := c.Read()
		if !ok {
			return nil
		}
		return v
	}
	return v
}

// Remove deletes the value stored under key and returns it.
func (s Store) Remove(key any) any {
	v, ok := s.storage.LoadAndDelete(key)
	if !ok {
		return nil
	}
	if c, ok := v.(*cell); ok {
		v, ok := c.Read()
		if !ok {
			return nil
		}
		return v
	}
	return v
}

// Init returns the value stored under key, creating it with new if needed.
func (s Store) Init(key any, new func() any) any {
beg:
	stored, ok := s.storage.Load(key)
	if ok {
		c, ok := stored.(*cell)
		if !ok {
			return stored
		}
		v, ok := c.Read()
		if ok {
			return v
		}
		if !s.storage.CompareAndDelete(key, stored) {
			goto beg
		}
	}
	newC := newCell()
retry:
	stored, ok = s.storage.LoadOrStore(key, newC)
	if ok {
		c, ok := stored.(*cell)
		if !ok {
			return stored
		}
		v, ok := c.Read()
		if ok {
			return v
		}
		if !s.storage.CompareAndSwap(key, stored, newC) {
			goto retry
		}
	}
	defer func() {
		if newC.Close() {
			return
		}
		s.storage.CompareAndDelete(key, newC)
	}()
	stored = new()
	newC.Put(stored)
	s.storage.CompareAndSwap(key, newC, stored)
	return stored
}

func newCell() *cell {
	cell := &cell{
		signal: make(chan struct{}),
	}
	return cell
}

type cellState int

const (
	cellInit cellState = iota
	cellComplete
	cellCanceled
)

type cell struct {
	value  any
	signal chan struct{}
	state  cellState
}

func (c *cell) Read() (any, bool) {
	<-c.signal
	if c.state == cellCanceled {
		return nil, false
	}
	return c.value, true
}

func (c *cell) Close() bool {
	if c.state == cellInit {
		c.state = cellCanceled
		close(c.signal)
		return false
	}
	return true
}

func (c *cell) Put(value any) {
	c.state = cellComplete
	c.value = value
	close(c.signal)
}
