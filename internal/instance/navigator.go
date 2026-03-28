// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
	"github.com/gammazero/deque"
)

func newNavigator[M any](
	inst *Instance[M],
	adapter path.Adapter[M],
	adapters path.Adapters,
	beam beam.Source[M],
	ctx context.Context,
	rerouted bool,
) *navigator[M] {
	return &navigator[M]{
		inst:     inst,
		adapter:  adapter,
		adapters: adapters,
		rerouted: rerouted,
		ctx:      ctx,
		model:    beam,
		first:    true,
	}
}

const historyLimit = 64

type navigator[M any] struct {
	inst     *Instance[M]
	adapter  path.Adapter[M]
	adapters path.Adapters
	model    beam.Source[M]
	mu       sync.Mutex
	history  deque.Deque[path.Location]
	ctx      context.Context
	seq      int
	rerouted bool
	first    bool
}

func (n *navigator[M]) newLink(a any) (core.Link, error) {
	m, ok := n.adapter.Assert(a)
	if ok {
		loc, err := n.adapter.Encode(m)
		if err != nil {
			return core.Link{}, err
		}
		return core.Link{
			Location: loc,
			On: func(ctx context.Context) {
				n.model.Update(ctx, *m)
			},
		}, nil
	}
	loc, err := n.adapters.Encode(a)
	if err != nil {
		return core.Link{}, err
	}
	return core.Link{
		Location: loc,
		On:       nil,
	}, nil
}

func (n *navigator[M]) init() {
	n.model.Sub(n.ctx, func(ctx context.Context, m M) bool {
		l, err := n.adapter.Encode(&m)
		if err != nil {
			slog.Error(
				"Path model encoding error on beam update",
				slog.String("error", err.Error()),
				slog.String("model", fmt.Sprintf("%+v", m)),
			)
			return false
		}
		n.push(ctx, l)
		return false
	})
}

func (n *navigator[M]) push(ctx context.Context, l path.Location) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.history.Len() != 0 && path.EqualLocation(n.history.Front(), l) {
		return
	}
	n.history.PushFront(l)
	if n.history.Len() > historyLimit {
		n.history.PopBack()
	}
	replace := false
	if n.first {
		n.first = false
		if !n.rerouted {
			return
		}
		replace = true
	}
	n.seq += 1
	seq := n.seq
	after, ok := ctex.AfterFrame(ctx)
	if !ok {
		n.call(l.Path(), seq, replace)
		return
	}
	after.RunAfter(nil, nil, func(b bool) {
		n.call(l.Path(), seq, replace)
	})
}

func (n *navigator[M]) call(path string, seq int, replace bool) {
	n.inst.CallCheck(
		func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			return seq == n.seq
		},
		&action.SetPath{Path: path, Replace: replace},
		nil,
		nil,
		action.CallParams{},
	)
}

func (n *navigator[M]) restore(l path.Location) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	for prev := range n.history.Iter() {
		if path.EqualLocation(prev, l) {
			goto found
		}
	}
	return false
found:
	m, ok := n.adapter.Decode(l)
	if !ok {
		slog.Error(
			"can't restore previous location, model decoding failed",
			slog.String("location", fmt.Sprintf("%+v", l)),
		)
		return false
	}
	n.model.Update(n.ctx, *m)
	return true
}
