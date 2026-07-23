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

package common

import (
	"context"
	"log/slog"
	"time"
)

type ctxKey int

const (
	KeyCore ctxKey = iota
	KeySession
	KeyFrame
	KeyHistoryReplace
)

func NewRenderCtx(system context.Context, user context.Context) RenderCtx {
	if rc, ok := system.(RenderCtx); ok {
		system = rc.system
	}
	return RenderCtx{system: system, user: user}
}

type RenderCtx struct {
	system context.Context
	user   context.Context
}

func (c RenderCtx) System() context.Context {
	return c.system
}

func (c RenderCtx) User() context.Context {
	return c.user
}

func (c RenderCtx) Deadline() (deadline time.Time, ok bool) {
	return c.system.Deadline()
}

func (c RenderCtx) Done() <-chan struct{} {
	return c.system.Done()
}

func (c RenderCtx) Err() error {
	return c.system.Err()
}

func (c RenderCtx) Value(key any) any {
	if _, ok := key.(ctxKey); ok {
		return c.system.Value(key)
	}
	return c.user.Value(key)
}

var _ context.Context = RenderCtx{}

func Logger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	lp, ok := ctx.Value(KeySession).(interface{ Logger() *slog.Logger })
	if !ok {
		return slog.Default()
	}
	return lp.Logger()
}
