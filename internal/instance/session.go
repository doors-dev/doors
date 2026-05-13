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
	//	"net/http"
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
)

type App interface {
	CSP() *common.CSP
	Conf() *common.Conf
	PathMaker() path.PathMaker
	RemoveSession(id string)
	ResourceRegistry() resources.Registry
}

type Session = *session

func NewSession(a App) Session {
	sess := &session{
		store:   ctex.NewStore(),
		id:      common.RandId(),
		app:     a,
		limiter: utils.NewLimiter(a.Conf().SessionInstanceLimit),
	}
	return sess
}

type session struct {
	instances  sync.Map
	killed     atomic.Bool
	mu         sync.Mutex
	store      ctex.Store
	id         string
	app        App
	limiter    utils.Limiter
	expireTime time.Time
	ttlTime    time.Time
	killTimer  *time.Timer
}

func (sess *session) App() core.App {
	return sess.app
}

func (sess *session) Expire(d time.Duration) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed.Load() {
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

func (sess Session) Store() ctex.Store {
	return sess.store
}

func (sess Session) Instance(loc path.Location) (Instance, bool) {
	if sess.killed.Load() {
		return nil, false
	}
	inst := newInstance(sess, loc)
	sess.instances.Store(inst.ID(), inst)
	toSuspend := sess.limiter.Add(inst.ID())
	if toSuspend == "" {
		return inst, !sess.killed.Load()
	}
	instToSuspend, loaded := sess.instances.LoadAndDelete(toSuspend)
	if !loaded {
		return inst, !sess.killed.Load()
	}
	instToSuspend.(Instance).end(common.EndCauseSuspend)
	return inst, !sess.killed.Load()
}

func (sess Session) InstanceCount() (n int) {
	sess.instances.Range(func(_, _ any) bool {
		n++
		return true
	})
	return
}

func (sess *session) Renew(w http.ResponseWriter) bool {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed.Load() {
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
	cookie := &http.Cookie{
		Name:     sess.app.PathMaker().SessionCookie(),
		Value:    sess.id,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
	}
	http.SetCookie(w, cookie)
	return !sess.killed.Load()
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
	if sess.killed.Load() {
		return nil, false
	}
	inst, ok := sess.instances.Load(id)
	if !ok {
		return nil, false
	}
	return inst.(Instance), !sess.killed.Load()
}

func (sess *session) Kill() {
	if !sess.killed.CompareAndSwap(false, true) {
		return
	}
	sess.mu.Lock()
	if sess.killTimer != nil {
		sess.killTimer.Stop()
	}
	sess.mu.Unlock()
	sess.app.RemoveSession(sess.id)
	sess.instances.Range(func(key, value any) bool {
		sess.instances.Delete(key)
		value.(Instance).end(common.EndCauseKilled)
		return true
	})
}
