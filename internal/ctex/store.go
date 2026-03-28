// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
