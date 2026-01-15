package renderer

import (
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/sh"
)

type Root struct {
	spawner *sh.Spawner
	prime   *common.Prime
}

func (r *Root) newId() uint64 {
	return r.prime.Gen()
}

func (r *Root) newThread() *sh.Thread {
	return r.spawner.NewThead()
}
