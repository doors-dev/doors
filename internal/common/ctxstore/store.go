// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package ctxstore

import (
	"context"
	"sync"
)

func NewStore(key any) *Store {
	return &Store{
		storage: make(map[any]any),
		key:     key,
	}
}

type Store struct {
	key     any
	storage map[any]any
	mu      sync.RWMutex
}

func (c *Store) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, c.key, c)
}

func (c *Store) Load(key any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.storage[key]
}

func (c *Store) Swap(key any, value any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v := c.storage[key]
	c.storage[key] = value
	return v
}

func (c *Store) Save(key any, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storage[key] = value
}

func (c *Store) Remove(key any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := c.storage[key]
	delete(c.storage, key)
	return v
}

func Swap(ctx context.Context, storeKey any, key any, value any) any {
	c := ctx.Value(storeKey).(*Store)
	return c.Swap(key, value)
}

func Load(ctx context.Context, storeKey any, key any) any {
	c := ctx.Value(storeKey).(*Store)
	return c.Load(key)
}

func Remove(ctx context.Context, storeKey any, key any) any {
	c := ctx.Value(storeKey).(*Store)
	return c.Remove(key)
}
