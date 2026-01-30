// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
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
	sess.setTTL()
	return sess

}

func (sess *Session) setTTL() {
	ttl := sess.router.Conf().SessionTTL
	if ttl == 0 {
		return
	}
	sess.ttl = time.AfterFunc(ttl, func() {
		sess.Kill()
	})
}

type Session struct {
	mu        sync.Mutex
	store     ctex.Store
	killed    bool
	id        string
	instances map[string]AnyInstance
	router    router
	limiter   *limiter
	expire    *time.Timer
	ttl       *time.Timer
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
	if len(sess.instances) == 0 && sess.ttl == nil {
		sess.killed = true
		sess.cleanup()
	}
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
	if sess.expire == nil {
		if d == 0 {
			return
		}
		sess.expire = time.AfterFunc(d, func() {
			sess.Kill()
		})
		return
	}
	if !sess.expire.Stop() {
		return
	}
	if d == 0 {
		sess.expire = nil
		if len(sess.instances) == 0 {
			sess.killed = true
			sess.cleanup()
		}
		return
	}
	sess.expire.Reset(d)
}

func (sess *Session) cleanup() {
	sess.router.RemoveSession(sess.id)
	if sess.expire != nil {
		sess.expire.Stop()
	}
	if sess.ttl != nil {
		sess.ttl.Stop()
	}
	for id := range sess.instances {
		sess.instances[id].end(common.EndCauseKilled)
	}
}
