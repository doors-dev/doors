package shredder

/*

type collector[T any] struct {
	collection *collection[T]
	collect    func(item T)
}

func (c *collector[T]) threadDone(*Thread) {
	c.collection.collect(c.collect)
}

func Collect[T any](
	thread *JoinedThread,
	f func(*Collector[T]),
	collect func(item T),
) bool {
	collection := newCollection[T]()
	collector := &collector[T]{
		collection: collection,
		collect:    collect,
	}
	run := func(t *Thread) {
		c := &Collector[T]{
			collection: collection,
			thread:     t,
		}
		t.addHead(collector)
		f(c)
	}
	if thread.write {
		thread.thread.Write(func(t *Thread) {
			run(t)
		})
	} else {
		thread.thread.Read(func(t *Thread) {
			run(t)
		})
	}
	return true
}

func newCollection[T any]() *collection[T] {
	return &collection[T]{
		branches: make([]*collection[T], 0),
		buffer:   make([]T, 0),
	}
}

type collection[T any] struct {
	branches []*collection[T]
	buffer   []T
}

func (c *collection[T]) collectBranches(f func(T)) {
	for _, branch := range c.branches {
		branch.collect(f)
	}
}

func (c *collection[T]) collect(f func(T)) {
	c.collectBranches(f)
	for _, item := range c.buffer {
		f(item)
	}
}

func (c *collection[T]) branch() *collection[T] {
	new := newCollection[T]()
	if len(c.buffer) != 0 {
		dump := newCollection[T]()
		items := c.buffer
		c.buffer = dump.buffer
		dump.buffer = items
		c.branches = append(c.branches, dump, new)
	} else {
		c.branches = append(c.branches, new)
	}
	return new
}

func (c *collection[T]) put(v T) {
	c.buffer = append(c.buffer, v)
}

type Collector[T any] struct {
	collection *collection[T]
	thread     *Thread
}

func (c *Collector[T]) Thread() *Thread {
	return c.thread
}

func (c *Collector[T]) Put(v T) {
	c.collection.put(v)
}

func (c *Collector[T]) Write(f func(*Collector[T]), join ...*JoinedThread) {
	c.execute(f, W(c.thread), join)
}

func (c *Collector[T]) Read(f func(*Collector[T]), join ...*JoinedThread) {
	c.execute(f, R(c.thread), join)
}

func (c *Collector[T]) execute(f func(*Collector[T]), self *JoinedThread, threads []*JoinedThread) {
	c.thread.mu.Lock()
	collection := c.collection.branch()
	c.thread.mu.Unlock()
	runMultiTask(c.thread.spawner, func(t *Thread) {
		var collector *Collector[T]
		if t != nil {
			collector = &Collector[T]{
				thread:     t,
				collection: collection,
			}
		}
		f(collector)
	},
		append([]*JoinedThread{self}, threads...),
	)
} */
