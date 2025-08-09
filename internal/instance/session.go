package instance

import (
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxstore"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
)

type ScriptOptions struct {
	Minify bool
	Gzip   bool
}

type router interface {
	ImportRegistry() *resources.Registry
	CSP() *common.CSP
	Adapters() map[string]path.AnyAdapter
	RemoveSession(string)
	Conf() *common.SystemConf
	Spawner(shredder.OnPanic) *shredder.Spawner
}

func NewSession(r router) *Session {
	sess := &Session{
		store:     ctxstore.NewStore(common.SessionStoreCtxKey),
		id:        common.RandId(),
		instances: make(map[string]AnyInstance),
		mu:        sync.Mutex{},
		router:    r,
		limiter:   newLimiter(r.Conf().InstanceGoroutineLimit),
	}
	sess.SetExpiration(r.Conf().SessionExpiration)
	return sess

}

type Session struct {
	store     *ctxstore.Store
	mu        sync.Mutex
	killed    bool
	id        string
	instances map[string]AnyInstance
	router    router
	limiter   *limiter
	expire    *time.Timer
}

func (sess *Session) getRouter() router {
	return sess.router
}
func (sess *Session) getStorage() *ctxstore.Store {
	return sess.store
}

func (sess *Session) AddInstance(inst AnyInstance) bool {
	sess.mu.Lock()
	if sess.killed {
		sess.mu.Unlock()
		return false
	}
	sess.instances[inst.Id()] = inst
	toSuspend := sess.limiter.add(inst.Id())
	sess.mu.Unlock()
	if toSuspend != "" {
		sess.instances[toSuspend].end(causeSuspend)
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
	if len(sess.instances) == 0 && sess.expire == nil {
		sess.killed = true
		sess.cleanup()
	}
}

func (sess *Session) Id() string {
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
	for id := range sess.instances {
		sess.instances[id].end(causeKilled)
	}
}
