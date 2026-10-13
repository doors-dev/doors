// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line retry_fragments.gox:1
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
package batch

import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

const retryBatchX = "x-batch"

const retryBatchY = "y-batch"

const retrySecondX = "x-second"

const retryThirdX = "x-third"

type retryState struct {
	entered chan struct{}
	proceed chan struct{}
	second chan struct{}
	third chan struct{}
}

func retryNewState() *retryState {
	return &retryState{
		entered: make(chan struct{}, 8),
		proceed: make(chan struct{}),
		second: make(chan struct{}, 8),
		third: make(chan struct{}, 8),
	}
}

type retryFragment struct {
	test.NoBeam
	s *retryState
	x doors.Door
	y doors.Door
}

func (f *retryFragment) batch(ctx context.Context) bool {
	<-f.x.Inner(ctx, retryBatchX)
	f.s.entered <- struct{}{}
	select {
	case <-f.s.proceed:
	case <-ctx.Done():
		return false
	}
	f.y.Inner(ctx, retryBatchY)
	return false
}

func (f *retryFragment) plain(ctx context.Context, content string, issued chan struct{}) bool {
	<-f.x.Inner(ctx, content)
	issued <- struct{}{}
	return false
}

//line retry_fragments.gox:73
func (f *retryFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line retry_fragments.gox:74
		__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line retry_fragments.gox:74
				__e = __c.Set("id", "x"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("x0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line retry_fragments.gox:75
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line retry_fragments.gox:75
				__e = __c.Set("id", "y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("y0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line retry_fragments.gox:76
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.batch(ctx)
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line retry_fragments.gox:81
				__e = __c.Set("id", "retry-batch"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("retry-batch"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line retry_fragments.gox:82
		__e = __c.Any(test.Button("retry-second", func(ctx context.Context) bool {
		return f.plain(ctx, retrySecondX, f.s.second)
	})); if __e != nil { return }
//line retry_fragments.gox:85
		__e = __c.Any(test.Button("retry-third", func(ctx context.Context) bool {
		return f.plain(ctx, retryThirdX, f.s.third)
	})); if __e != nil { return }
	return })
//line retry_fragments.gox:88
}

var _ gox.Comp = (*retryFragment)(nil)
