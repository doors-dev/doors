package door2

import (
	"context"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
)

type Root struct {
	ctx     context.Context
	cancel  context.CancelFunc
	prime   *common.Prime
	spawner sh.Spawner
}

func (r *Root) Spawner() sh.Spawner {
	return r.spawner
}

func (r *Root) getContext() context.Context {
	return r.ctx
}

func (r *Root) getRoot() *Root {
	return r
}

func (r *Root) newId() uint64 {
	return r.prime.Gen()
}

