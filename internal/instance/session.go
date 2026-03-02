// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"net/http"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
)

type ScriptOptions struct {
	Minify bool
	Gzip   bool
}

type router interface {
	ResourceRegistry() *resources.Registry
	CSP() *common.CSP
	Adapters() map[string]path.AnyAdapter
	RemoveSession(string)
	Conf() *common.SystemConf
	License() license.License
}

func NewSession(r router) *Session {
	sess := &Session{
		store:     ctex.NewStore(),
		id:        common.RandId(),
		instances: make(map[string]AnyInstance),
		mu:        sync.Mutex{},
		router:    r,
		limiter:   newLimiter(r.Conf().SessionInstanceLimit),
	}
	return sess
}

type Session struct {
	mu         sync.Mutex
	store      ctex.Store
	killed     bool
	id         string
	instances  map[string]AnyInstance
	router     router
	limiter    *limiter
	expireTime time.Time
	ttlTime    time.Time
	killTimer  *time.Timer
}

func (sess *Session) Store() ctex.Store {
	return sess.store
}

func (sess *Session) Renew(w http.ResponseWriter) bool {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	ttl := sess.router.Conf().SessionTTL
	sess.ttlTime = time.Now().Add(ttl)
	if !sess.resetKillTimer() {
		return false
	}
	maxAge := sess.untillKill()
	if maxAge < sess.router.Conf().RequestTimeout {
		return false
	}
	cookie := &http.Cookie{
		Name:     "d0-r",
		Value:    sess.id,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge.Seconds()),
	}
	http.SetCookie(w, cookie)
	return true
}

func (sess *Session) resetKillTimer() bool {
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

func (sess *Session) killTime() time.Time {
	if sess.expireTime.IsZero() {
		return sess.ttlTime
	}
	if sess.ttlTime.Before(sess.expireTime) {
		return sess.ttlTime
	}
	return sess.expireTime
}

func (sess *Session) untillKill() time.Duration {
	return time.Until(sess.killTime())
}

func (sess *Session) AddInstance(inst AnyInstance) bool {
	sess.mu.Lock()
	if sess.killed {
		sess.mu.Unlock()
		return false
	}
	sess.instances[inst.ID()] = inst
	toSuspend := sess.limiter.add(inst.ID())
	sess.mu.Unlock()
	if toSuspend != "" {
		sess.instances[toSuspend].end(common.EndCauseSuspend)
	}
	return true
}

func (sess *Session) removeInstance(id string) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed {
		return
	}
	sess.limiter.delete(id)
	delete(sess.instances, id)
}

func (sess *Session) ID() string {
	return sess.id
}

func (sess *Session) GetInstance(id string) (AnyInstance, bool) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed {
		return nil, false
	}
	inst, ok := sess.instances[id]
	return inst, ok
}

func (sess *Session) Kill() {
	sess.mu.Lock()
	if sess.killed {
		sess.mu.Unlock()
		return
	}
	sess.killed = true
	sess.mu.Unlock()
	sess.cleanup()
}

func (sess *Session) SetExpiration(d time.Duration) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.killed {
		return
	}
	sess.expireTime = time.Now().Add(d)
	sess.resetKillTimer()
}

func (sess *Session) cleanup() {
	sess.router.RemoveSession(sess.id)
	if sess.killTimer != nil {
		sess.killTimer.Stop()
	}
	for id := range sess.instances {
		sess.instances[id].end(common.EndCauseKilled)
	}
}
