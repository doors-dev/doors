package node

import (
	"context"
	"errors"

	"github.com/doors-dev/doors/internal/common/ctxwg"
)

func newCommit(ctx context.Context) *commit {
	c := &commit{
		ch: make(chan error, 1),
	}
	if ctx != nil {
		c.ctx = ctx
		c.done = ctxwg.Add(ctx)
	}
	return c
}

type commit struct {
	ctx  context.Context
	done func()
	ch   chan error
}

func (c *commit) result(err error) {
	c.ch <- err
	close(c.ch)
	if c.done != nil {
		c.done()
	}
}
func (c *commit) owerwrite() {
	c.result(errors.New("operation overwritten"))
}

func (c *commit) suspend() {
	close(c.ch)
	if c.done != nil {
		c.done()
	}
}
