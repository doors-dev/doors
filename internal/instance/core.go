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
	relocate(any) error
	end()
	include() bool
}

func newCore[M any](inst coreInstance[M], connector *connector, spawner *shredder.Spawner) *core[M] {
	return &core[M]{
		instance:     inst,
		gen:          common.NewPrima(),
		hooksMu:      sync.Mutex{},
		hooks:        make(map[uint64]map[uint64]*node.NodeHook),
		spawner:      spawner,
		connector:    connector,
		cspCollector: inst.getSession().getRouter().CSP().NewCollector(),
	}
}

type Core interface {
	InlineNonce() (string, bool)
	CSPCollector() (*common.CSPCollector, bool)
	ImportRegistry() *resources.Registry
	SessionId() string
	Relocate(context.Context, any) error
	Include() bool
	TTL() time.Duration
	Id() string
	NewId() uint64
	NewLink(context.Context, any) (*Link, error)
	SessionExpire(d time.Duration)
	SessionEnd()
	End()
}

type core[M any] struct {
	instance         coreInstance[M]
	gen              *common.Prima
	hooksMu          sync.Mutex
	hooks            map[uint64]map[uint64]*node.NodeHook
	root             node.Core
	cinema           *node.Cinema
	connector        *connector
	spawner          *shredder.Spawner
	cspCollectorUsed atomic.Bool
	cspCollector     *common.CSPCollector
	nonce            string
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

func (c *core[M]) Relocate(ctx context.Context, model any) error {

	if ctx.Err() != nil {
		return errors.New("context not active")
	}
	return c.instance.relocate(model)
}

func (c *core[M]) Sync(ctx context.Context, screenId uint64, seq uint, collector *shredder.Collector[func()]) {
	c.cinema.InitSync(ctx, screenId, seq, collector)
}

func (c *core[M]) NewId() uint64 {
	return c.gen.Gen()
}

func (c *core[M]) SessionEnd() {
	c.instance.getSession().Kill()
}

func (c *core[M]) End() {
	c.instance.end()
}

func (c *core[M]) SessionExpire(d time.Duration) {
	c.instance.getSession().SetExpiration(d)
}

func (c *core[M]) TTL() time.Duration {
	return c.connector.ttl
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
func (c *core[M]) kill(suspend bool) {
	c.connector.Kill(suspend)
	if c.root != nil {
		c.root.Kill()
	}
	c.spawner.Kill()
}

func (c *core[M]) Thread() *shredder.Thread {
	return c.spawner.NewThead()
}

func (c *core[M]) Call(caller node.Caller) {
	c.connector.Call(caller)
}

func (c *core[M]) Setup(root node.Core, cinema *node.Cinema, ctx context.Context) {
	c.root = root
	c.cinema = cinema
	c.instance.setupPathSync(ctx)
}

func (c *core[M]) serve(w http.ResponseWriter, content templ.Component, code int) error {
	n := node.Node{}
	ch := make(chan error, 0)
	rm := common.NewRenderMap()
	thread := c.spawner.NewThead()
	thread.Write(func(t *shredder.Thread) {
		if t == nil {
			return
		}
		ctx := context.WithValue(context.Background(), common.InstanceCtxKey, c)
		ctx = c.instance.getSession().getStorage().Inject(ctx)
		ctx = context.WithValue(ctx, common.ThreadCtxKey, t)
		ctx = context.WithValue(ctx, common.RenderMapCtxKey, rm)
		ctx = context.WithValue(ctx, common.AdaptersCtxKey, c.instance.getSession().getRouter().Adapters())
		n.Update(ctx, content)
		n.Render(ctx, nil)
	})
	thread.Write(func(t *shredder.Thread) {
		if t == nil {
			ch <- errors.New("instance killed before render")
		}
		close(ch)
	})
	err, ok := <-ch
	if ok {
		return err
	}
	if c.cspCollector != nil {
		c.cspCollectorUsed.Store(true)
		c.cspCollector.ScriptHash(c.instance.getSession().getRouter().ImportRegistry().MainScript().Hash())
		c.cspCollector.StyleHash(c.instance.getSession().getRouter().ImportRegistry().MainStyle().Hash())
		header := c.cspCollector.Generate()
		w.Header().Add("Content-Security-Policy", header)
	}
	defer rm.Destroy()
	if c.instance.getSession().getRouter().Gzip() {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(code)
		writer := gzip.NewWriter(w)
		err = rm.Render(writer, c.root.Id())
		if err != nil {
			return err
		}
		return writer.Close()
	}
	w.WriteHeader(code)
	err = rm.Render(w, c.root.Id())
	return err
}
