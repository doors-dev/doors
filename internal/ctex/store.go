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
	return &store{
		storage: make(map[any]any),
	}
}

type store struct {
	storage map[any]any
	mu      sync.RWMutex
}

// Load returns the value stored under key or nil if key is absent.
func (c Store) Load(key any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.storage[key]
}

// Save stores value under key and returns the previous value.
func (c Store) Save(key any, value any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := c.storage[key]
	c.storage[key] = value
	return v
}

// Remove deletes the value stored under key and returns it.
func (c Store) Remove(key any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := c.storage[key]
	delete(c.storage, key)
	return v
}

// Init returns the value stored under key, creating it with new if needed.
func (c Store) Init(key any, new func() any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.storage[key]; ok {
		return v
	}
	v := new()
	c.storage[key] = v
	return v
}

/*

func StoreInit(ctx context.Context, storeKey any, key any, new func() any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Init(key, new)
}

// SessionSave stores a key/value in session-scoped storage shared by all
// instances in the session. Returns the previous value under the key.
func StoreSave(ctx context.Context, storeKey any, key any, value any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Save(key, value)
}

func StoreLoad(ctx context.Context, storeKey any, key any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Load(key)
}

func StoreRemove(ctx context.Context, storeKey any, key any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Remove(key)
} */
