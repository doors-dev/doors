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

package utils

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
)

type Navigator = *navigator

func NewNavigator(
	inst core.Instance,
	ctx context.Context,
) Navigator {
	initial := inst.Location().Get()
	n := &navigator{
		inst:   inst,
		inital: &initial,
		ctx:    ctx,
	}
	return n
}

type navigator struct {
	inst   core.Instance
	inital *path.Location
	ctx    context.Context
	model  beam.Source[path.Location]
	seq    atomic.Int32
}

func (n *navigator) Sync() {
	n.inst.Location().Sub(n.ctx, func(ctx context.Context, l path.Location) bool {
		if n.inital != nil {
			if !path.EqualLocation(l, *n.inital) {
				n.push(ctx, l, true)
			}
			n.inital = nil
			return false
		}
		n.push(ctx, l, false)
		return false
	})
}

func (n *navigator) Update(l path.Location) bool {
	n.inst.Location().Update(n.ctx, l)
	return true
}

func (n *navigator) push(ctx context.Context, l path.Location, replace bool) {
	seq := n.seq.Add(1)
	after, ok := ctex.AfterFrame(ctx)
	if !ok {
		n.call(l.String(), seq, replace)
		return
	}
	after.RunAfter(nil, nil, func(b bool) {
		n.call(l.String(), seq, replace)
	})
}

func (n *navigator) call(path string, seq int32, replace bool) {
	n.inst.UserCall(
		context.Background(),
		func() bool {
			return seq == n.seq.Load()
		},
		&action.SetPath{Path: path, Replace: replace},
		nil,
		nil,
		action.CallParams{},
	)
}
