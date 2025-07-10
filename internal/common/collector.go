package common

import "sync"

func NewFuncCollector() *FuncCollector {
	return &FuncCollector{
		collection: make([]func(), 0),
	}
}

type FuncCollector struct {
	mu         sync.Mutex
	collection []func()
}

func (c *FuncCollector) Add(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, f)
}

func (c *FuncCollector) Apply() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, f := range c.collection {
		f()
	}
	c.collection = make([]func(), 0)
}
