package instance

import (
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxstore"
	"github.com/doors-dev/doors/internal/node"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
)

type coreSession interface {
	getStorage() *ctxstore.Store
	SetExpiration(time.Duration)
	Kill()
	Id() string
	getRouter() router
}

type coreInstance[M any] interface {
	getSession() coreSession
	Id() string
	setupPathSync(context.Context)
	newLink(context.Context, any) (*Link, error)
	end(endCause)
	include() bool
	conf() *common.SystemConf
	OnPanic(error)
}

func newCore[M any](inst coreInstance[M], solitaire *solitaire, spawner *shredder.Spawner) *core[M] {
	return &core[M]{
		instance:     inst,
		store:        ctxstore.NewStore(common.InstanceStoreCtxKey),
		gen:          common.NewPrima(),
		hooksMu:      sync.Mutex{},
		hooks:        make(map[uint64]map[uint64]*node.NodeHook),
		spawner:      spawner,
		solitaire:    solitaire,
		cspCollector: inst.getSession().getRouter().CSP().NewCollector(),
	}
}

type Core interface {
	Thread() *shredder.Thread
	InlineNonce() (string, bool)
	CSPCollector() (*common.CSPCollector, bool)
	ImportRegistry() *resources.Registry
	SessionId() string
	Include() bool
	ClientConf() *common.ClientConf
	Id() string
	NewId() uint64
	Cinema() *node.Cinema
	NewLink(context.Context, any) (*Link, error)
	SessionExpire(d time.Duration)
	SessionEnd()
	Call(call common.Call)
	End()
}

type core[M any] struct {
	instance         coreInstance[M]
	store            *ctxstore.Store
	gen              *common.Primea
	hooksMu          sync.Mutex
	hooks            map[uint64]map[uint64]*node.NodeHook
	root             *node.Root
	solitaire        *solitaire
	spawner          *shredder.Spawner
	cspCollectorUsed atomic.Bool
	cspCollector     *common.CSPCollector
	nonce            string
}

func (c *core[M]) OnPanic(err error) {
	c.instance.OnPanic(err)
}

func (c *core[M]) SleepTimeout() time.Duration {
	return c.instance.conf().ClientHiddenSleepTimer
}

func (c *core[M]) ClientConf() *common.ClientConf {
	return common.GetClientConf(c.instance.conf())
}

func (c *core[M]) InlineNonce() (string, bool) {
	nonce := c.cspCollector.Nonce()
	return nonce, nonce != ""
}

func (c *core[M]) CSPCollector() (*common.CSPCollector, bool) {
	if c.cspCollector == nil {
		return nil, true
	}
	if !c.cspCollectorUsed.CompareAndSwap(false, true) {
		return nil, false
	}
	return c.cspCollector, true
}

func (c *core[M]) ImportRegistry() *resources.Registry {
	return c.instance.getSession().getRouter().ImportRegistry()
}

func (c *core[M]) SessionId() string {
	return c.instance.getSession().Id()
}

func (c *core[M]) Cinema() *node.Cinema {
	return c.root.Cinema()
}

func (c *core[M]) NewId() uint64 {
	return c.gen.Gen()
}

func (c *core[M]) SessionEnd() {
	c.instance.getSession().Kill()
}

func (c *core[M]) End() {
	c.instance.end(causeKilled)
}

func (c *core[M]) SessionExpire(d time.Duration) {
	c.instance.getSession().SetExpiration(d)
}

func (c *core[M]) Id() string {
	return c.instance.Id()
}

func (c *core[M]) Include() bool {
	return c.instance.include()
}

func (c *core[M]) NewLink(ctx context.Context, m any) (*Link, error) {
	return c.instance.newLink(ctx, m)
}
func (c *core[M]) end(cause endCause) {
	c.solitaire.End(cause)
	if c.root != nil {
		c.root.Kill()
	}
	c.spawner.Kill()
}

func (c *core[M]) Thread() *shredder.Thread {
	return c.spawner.NewThead()
}

func (c *core[M]) Call(call common.Call) {
	c.solitaire.Call(call)
}

func (c *core[M]) serve(w http.ResponseWriter, content templ.Component, code int) error {
	ctx := context.WithValue(context.Background(), common.InstanceCtxKey, c)
	ctx = c.store.Inject(ctx)
	ctx = context.WithValue(ctx, common.AdaptersCtxKey, c.instance.getSession().getRouter().Adapters())

	c.root = node.NewRoot(ctx, c)

	ch := c.root.Render(content)
	render, ok := <-ch
	if !ok {
		return errors.New("instance killed before render")
	}
	if c.cspCollector != nil {
		c.cspCollectorUsed.Store(true)
		c.cspCollector.ScriptHash(c.instance.getSession().getRouter().ImportRegistry().MainScript().Hash())
		c.cspCollector.StyleHash(c.instance.getSession().getRouter().ImportRegistry().MainStyle().Hash())
		header := c.cspCollector.Generate()
		w.Header().Add("Content-Security-Policy", header)
	}
	gz := !c.instance.conf().ServerDisableGzip

	if gz {
		w.Header().Set("Content-Encoding", "gzip")
	}

	if render.Err() != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(code)
	}

	var err error
	if gz {
		writer := gzip.NewWriter(w)
		err = render.Write(writer)
		if err == nil {
			err = writer.Close()
		}
	} else {
		err = render.Write(w)
	}
	if err != nil {
		return err
	}
	c.instance.setupPathSync(c.root.Ctx())
	return nil
}
