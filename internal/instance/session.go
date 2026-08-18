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
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/instance/utils"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type App interface {
	CSP() *common.CSP
	Conf() *common.Conf
	SetCookies(w http.ResponseWriter, id string, maxAge time.Duration)
	PathMaker() path.PathMaker
	RemoveSession(id string)
	ResourceRegistry() resources.Registry
	Logger() *slog.Logger
	InstanceCreated()
	InstanceDeleted()
	Draining() bool
	PrinterMiddleware() func(next gox.Printer) gox.Printer
}

type Session = *session

func NewSession(a App, id string) Session {
	ctx, cancel := context.WithCancel(context.Background())
	sess := &session{
		store:   ctex.NewStore(),
		id:      id,
		app:     a,
		limiter: utils.NewLimiter(a.Conf().SessionInstanceLimit),
		cancel:  cancel,
	}
	sess.ctx = context.WithValue(ctx, common.KeySession, sess)
	now := time.Now()
	sess.renewed.Store(now.UnixMilli())
	sess.ttlTime = now.Add(a.Conf().SessionTTL)
	sess.resetKillTimer()
	return sess
}

type session struct {
	renewed    atomic.Int64
	instances  sync.Map
	mu         sync.Mutex
	store      ctex.Store
	id         string
	app        App
	limiter    utils.Limiter
	expireTime time.Time
	ttlTime    time.Time
	killTimer  *time.Timer
	ctx        context.Context
	cancel     context.CancelFunc
}

func (sess *session) Logger() *slog.Logger {
	return sess.app.Logger()
}

func (sess *session) killed() bool {
	return sess.ctx.Err() != nil
}

func (sess *session) Context() context.Context {
	return sess.ctx
}

func (sess *session) Store() ctex.Store {
	return sess.store
}

func (sess *session) App() core.App {
	return sess.app
}

func (sess *session) Expire(d time.Duration) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed() {
		return
	}
	if d == 0 {
		sess.expireTime = time.Time{}
	} else {
		sess.expireTime = time.Now().Add(d)
	}
	sess.resetKillTimer()
}

func (sess Session) ID() string {
	return sess.id
}

func (sess *session) LastSeen() time.Time {
	return time.UnixMilli(sess.renewed.Load())
}

func (sess Session) Instance(loc path.Location) (Instance, bool) {
	if sess.killed() {
		return nil, false
	}
	inst := newInstance(sess, loc)
	sess.instances.Store(inst.ID(), inst)
	toSuspend := sess.limiter.Add(inst.ID())
	if toSuspend == "" {
		return inst, !sess.killed()
	}
	instToSuspend, loaded := sess.instances.LoadAndDelete(toSuspend)
	if !loaded {
		return inst, !sess.killed()
	}
	instToSuspend.(Instance).end(common.EndCauseSuspend)
	return inst, !sess.killed()
}

func (sess Session) InstanceCount() (n int) {
	sess.instances.Range(func(_, _ any) bool {
		n++
		return true
	})
	return
}

const UPDATE_PERIOD = time.Second * 5

func (sess *session) Renew(w http.ResponseWriter) bool {
	now := time.Now().UnixMilli()
	before := sess.renewed.Load()
	ping := sess.App().Conf().SolitaireRollTime
	if now-before < ping.Milliseconds()/2 {
		return !sess.killed()
	}
	if !sess.renewed.CompareAndSwap(before, now) {
		return !sess.killed()
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed() {
		return false
	}
	ttl := sess.app.Conf().SessionTTL
	sess.ttlTime = time.Now().Add(ttl)
	if !sess.resetKillTimer() {
		return false
	}
	maxAge := sess.untillKill()
	if maxAge < sess.app.Conf().RequestTimeout {
		return false
	}
	sess.app.SetCookies(w, sess.id, maxAge)
	return !sess.killed()
}

func (sess *session) resetKillTimer() bool {
	if sess.killTimer != nil {
		if !sess.killTimer.Stop() {
			return false
		}
	}
	ttl := sess.untillKill()
	if ttl <= 0 {
		return false
	}
	if sess.killTimer != nil {
		sess.killTimer.Reset(ttl)
		return true
	}
	sess.killTimer = time.AfterFunc(ttl, func() {
		sess.Kill()
	})
	return true
}

func (sess *session) killTime() time.Time {
	if sess.expireTime.IsZero() {
		return sess.ttlTime
	}
	if sess.ttlTime.Before(sess.expireTime) {
		return sess.ttlTime
	}
	return sess.expireTime
}

func (sess *session) untillKill() time.Duration {
	return time.Until(sess.killTime())
}

func (sess *session) removeInstance(id string) {
	sess.limiter.Delete(id)
	sess.instances.Delete(id)
}

func (sess *session) GetInstance(id string) (Instance, bool) {
	if sess.killed() {
		return nil, false
	}
	inst, ok := sess.instances.Load(id)
	if !ok {
		return nil, false
	}
	return inst.(Instance), !sess.killed()
}

func (sess *session) Kill() {
	sess.cancel()
	sess.mu.Lock()
	if sess.killTimer != nil {
		sess.killTimer.Stop()
		sess.killTimer = nil
	}
	sess.mu.Unlock()
	sess.app.RemoveSession(sess.id)
}
