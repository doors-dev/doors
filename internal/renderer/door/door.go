package door

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/gox"
)

type door struct {
	state atomic.Value
}

func (n *door) Proxy(ctx context.Context, cur gox.Cursor, elem gox.Elem) error {
	newState := &proxyState{
		ctx:     ctx,
		element: elem,
	}
	n.takeover(newState)
	return cur.Job(newState)
}

func (n *door) Job(ctx context.Context) gox.Job {
	newState := &jobState{
		ctx: ctx,
	}
	n.takeover(newState)
	return newState
}

func (n *door) update(ctx context.Context, content any) {
	newState := &updateState{
		ctx:     ctx,
		content: content,
	}
	n.takeover(newState)
}

func (n *door) takeover(newState state) {
	value := n.state.Load()
	var prevState state
	if value != nil {
		prevState = value.(state)
	} else {
		u := updateState{
			ctx:     context.Background(),
			content: nil,
		}
		u.initFinished()
		u.readyForTakeover()
		prevState = &u
	}
	prevState.afterInit(func(bool) {
		newState.takeover(prevState)
	})
}
