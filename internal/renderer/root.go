package renderer

import (
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder2"
)

type Root struct {
	spawner *shredder2.Spawner
	prime   *common.Prime
}

func (r *Root) newId() uint64 {
	return r.prime.Gen()
}

func (r *Root) newThread() *shredder2.Thread {
	return r.spawner.NewThead()
}
