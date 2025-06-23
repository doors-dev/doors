package ctxstore

import (
	"context"
	"sync"
)

func NewStore() *Store {
	return &Store{
		storage: make(map[any]any),
	}
}

type Store struct {
	storage map[any]any
	mu      sync.RWMutex
}


type storeKeyType struct{}

var storeKey = storeKeyType{}

func  (c *Store) Inject(ctx context.Context) context.Context {
    return context.WithValue(ctx, storeKey, c)
}

func (c *Store) Load(key any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, _ := c.storage[key]
	return v
}

func (c *Store) Save(key any, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storage[key] = value
}

func Save(ctx context.Context, key any, value any) bool {
    c, ok := ctx.Value(storeKey).(*Store)
    if !ok {
        return true
    }
    c.Save(key, value)
    return false
}

func Load(ctx context.Context, key any) any {
    c, ok := ctx.Value(storeKey).(*Store)
    if !ok {
        return nil
    }
    return c.Load(key)
}
