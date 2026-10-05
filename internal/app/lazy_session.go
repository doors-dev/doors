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

package app

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/instance"
)

func getSession(ctx context.Context) (instance.Session, bool) {
	switch s := ctx.Value(common.KeySession).(type) {
	case instance.Session:
		return s, true
	case *lazySession:
		return s.get(), true
	default:
		return nil, false
	}
}

type lazySession struct {
	id      string
	r       *http.Request
	app     *app
	once    sync.Once
	session instance.Session
}

var _ core.Session = (*lazySession)(nil)

func (l *lazySession) Context() context.Context {
	return l.get().Context()
}

func (l *lazySession) Expire(d time.Duration) {
	l.get().Expire(d)
}

func (l *lazySession) ID() string {
	return l.get().ID()
}

func (l *lazySession) Kill() {
	l.get().Kill()
}

func (l *lazySession) LastSeen() time.Time {
	return l.get().LastSeen()
}

func (l *lazySession) Store() ctex.Store {
	return l.get().Store()
}

func (l *lazySession) App() core.App {
	return l.app
}

func (l *lazySession) Logger() *slog.Logger {
	return l.app.logger
}

func (l *lazySession) get() instance.Session {
	l.once.Do(func() {
		l.session = l.app.newSession(l.r, l.id)
	})
	return l.session
}
