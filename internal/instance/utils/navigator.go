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
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/actions"
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
		target: inst.NewID(),
	}
	return n
}

type navigator struct {
	inst   core.Instance
	target uint64
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
		replace, _ := ctx.Value(common.KeyHistoryReplace).(bool)
		n.push(ctx, l, replace)
		return false
	})
}

func (n *navigator) Update(l path.Location) bool {
	n.inst.Location().Update(n.ctx, l)
	return true
}

func (n *navigator) push(ctx context.Context, l path.Location, replace bool) {
	seq := n.seq.Add(1)
	frames := ctex.GetFrames(ctx)
	group := frames.Group()
	batch := frames.Batch()
	after, ok := frames.AfterFrame()
	if !ok {
		n.call(l.String(), seq, replace, group, batch)
		return
	}
	after.After().Run(nil, nil, func(b bool) {
		n.call(l.String(), seq, replace, group, batch)
	})
}

func (n *navigator) call(path string, seq int32, replace bool, group uint64, batch uint64) {
	n.inst.UserCallCheck(
		group,
		batch,
		n.target,
		func() bool {
			return seq == n.seq.Load()
		},
		&actions.SetPath{Path: path, Replace: replace},
		nil,
		nil,
		actions.CallParams{},
	)
}
