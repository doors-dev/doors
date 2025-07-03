package common

import "sync"

func NewCollector() *Collector {
	return &Collector{
		collection: make([]func(), 0),
	}
}

type Collector struct {
	mu         sync.Mutex
	collection []func()
}

func (c *Collector) Push(f func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collection = append(c.collection, f)
}

func (c *Collector) Collect() []func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.collection
}

func (c *Collector) Apply() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, f := range c.collection {
		f()
	}
	c.collection = make([]func(), 0)
}
