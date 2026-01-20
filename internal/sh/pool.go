package sh

import (
	"runtime"
	"sync"
)

var pool *sync.Pool

func init() {
	pool = &sync.Pool{
		New: func() any {
			exec := make(chan func(), 1)
			ref := &exec
			go func() {
				for task := range exec {
					task()
					pool.Put(ref)
				}
			}()
			runtime.AddCleanup(ref, func(ch chan func()) {
				close(exec)
			}, exec)
			return ref
		},
	}
}

func Go(f func()) {
	ref := pool.Get().(*chan func())
	exec := *ref
	exec <- f
}

