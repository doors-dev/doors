// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package ctex

import (
	"context"
	"sync"
)

type Store = *store

func NewStore() Store {
	return &store{
		storage: make(map[any]any),
	}
}

type store struct {
	storage map[any]any
	mu      sync.RWMutex
}

func (c *store) Inject(ctx context.Context, key any) context.Context {
	return context.WithValue(ctx, key, c)
}

func (c *store) Load(key any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.storage[key]
}

func (c *store) Swap(key any, value any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v := c.storage[key]
	c.storage[key] = value
	return v
}

func (c *store) Save(key any, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storage[key] = value
}

func (c *store) Remove(key any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := c.storage[key]
	delete(c.storage, key)
	return v
}

func (c *store) Init(key any, new func() any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.storage[key]; ok {
		return v
	}
	v := new()
	c.storage[key] = v
	return v
}

func StoreInit(ctx context.Context, storeKey any, key any, new func() any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Init(key, new)
}

func StoreSwap(ctx context.Context, storeKey any, key any, value any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Swap(key, value)
}

func StoreLoad(ctx context.Context, storeKey any, key any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Load(key)
}

func StoreRemove(ctx context.Context, storeKey any, key any) any {
	c := ctx.Value(storeKey).(*store)
	return c.Remove(key)
}
