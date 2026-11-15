// Managed by GoX v0.4.0

//line checkpoint.gox:1
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


package attr

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type checkpointLog struct {
	mu      sync.Mutex
	entries []string
}

func (l *checkpointLog) add(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

func (l *checkpointLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.entries)
}

type checkpointFragment struct {
	log *checkpointLog
	test.NoBeam
}

//line checkpoint.gox:51
func (f *checkpointFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line checkpoint.gox:53
		cp := &doors.ScopeCheckpoint{}
	blocking := &doors.ScopeBlocking{}

//line checkpoint.gox:56
		__e = __c.Any(f.button("o1", cp.Scope(false), 600*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:57
		__e = __c.Any(f.button("o2", cp.Scope(false), 900*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:58
		__e = __c.Any(f.button("of", cp.Scope(false), 0)); if __e != nil { return }
//line checkpoint.gox:59
		__e = __c.Any(f.button("c1", cp.Scope(true), 400*time.Millisecond)); if __e != nil { return }
//line checkpoint.gox:60
		__e = __c.Any(f.button("c2", cp.Scope(true), 0)); if __e != nil { return }
//line checkpoint.gox:61
		__e = __c.Any(f.button("c3", cp.Scope(true), 0)); if __e != nil { return }
//line checkpoint.gox:62
		__e = __c.Any(f.button("cb", blocking.And(cp.Scope(true)), 400*time.Millisecond)); if __e != nil { return }
	return })
//line checkpoint.gox:63
}

//line checkpoint.gox:65
func (f *checkpointFragment) button(id string, scope doors.Scopes, delay time.Duration) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line checkpoint.gox:66
			__e = __c.Set("id", id); if __e != nil { return }
//line checkpoint.gox:66
			__e = __c.Modify(doors.A(ctx, f.handler(id, scope, delay))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line checkpoint.gox:66
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line checkpoint.gox:67
}

func (f *checkpointFragment) handler(id string, scope doors.Scopes, delay time.Duration) doors.Attr {
	return doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.log.add("start:" + id)
			<-time.After(delay)
			f.log.add("end:" + id)
			return false
		},
	}
}
