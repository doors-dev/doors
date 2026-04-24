// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
)

func newNavigator[M any](
	inst *Instance[M],
	adapter path.Adapter[M],
	adapters path.Adapters,
	beam *beam.SourceBeam[M],
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

type navigator[M any] struct {
	inst     *Instance[M]
	adapter  path.Adapter[M]
	adapters path.Adapters
	model    *beam.SourceBeam[M]
	mu       sync.Mutex
	ctx      context.Context
	seq      int
	rerouted bool
	first    bool
}

func (n *navigator[M]) restore(l path.Location) bool {
	m, ok := n.adapter.Decode(l)
	if !ok {
		return false
	}
	n.model.Update(n.ctx, *m)
	return true
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
				"error",
				err,
				"model",
				fmt.Sprintf("%+v", m),
			)
			return false
		}
		n.push(ctx, l)
		return false
	})
}

func (n *navigator[M]) push(ctx context.Context, l path.Location) {
	n.mu.Lock()
	replace := false
	if n.first {
		n.first = false
		if !n.rerouted {
			n.mu.Unlock()
			return
		}
		replace = true
	}
	n.seq += 1
	seq := n.seq
	after, ok := ctex.AfterFrame(ctx)
	if !ok {
		n.mu.Unlock()
		n.call(l.String(), seq, replace)
		return
	}
	n.mu.Unlock()
	after.RunAfter(nil, nil, func(b bool) {
		n.call(l.String(), seq, replace)
	})
}

func (n *navigator[M]) call(path string, seq int, replace bool) {
	n.inst.UserCall(
		context.Background(),
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
