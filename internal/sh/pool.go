package sh

import (
	"runtime"
	"sync"

	"github.com/doors-dev/doors/internal/common"
)

var pool *sync.Pool

type pooltask struct {
	run func()
	done func(error)
}

func (t pooltask) exec() {
	err := common.Catch(t.run)
	if t.done == nil {
		return
	}
	t.done(err)
}

func init() {
	pool = &sync.Pool{
		New: func() any {
			exec := make(chan pooltask, 0)
			ref := &exec
			go func() {
				for task := range exec {
					task.exec()
					pool.Put(ref)
				}
			}()
			runtime.AddCleanup(ref, func(ch chan pooltask) {
				close(ch)
			}, exec)
			return ref
		},
	}
}



func Go(run func(), done func(error)) {
	ref := pool.Get().(*chan pooltask)
	exec := *ref
	exec <- pooltask{
		run: run,
		done: done,
	}
}

