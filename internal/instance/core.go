// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxstore"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front/action"
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
	end(endCause)
	conf() *common.SystemConf
	OnPanic(error)
}

func newCore[M any](inst coreInstance[M], solitaire *solitaire, spawner *shredder.Spawner, navigator *navigator[M]) *core[M] {
	return &core[M]{
		instance:     inst,
		store:        ctxstore.NewStore(common.CtxKeyInstanceStore),
		gen:          common.NewPrima(),
		hooksMu:      sync.Mutex{},
		hooks:        make(map[uint64]map[uint64]*door.DoorHook),
		spawner:      spawner,
		solitaire:    solitaire,
		cspCollector: inst.getSession().getRouter().CSP().NewCollector(),
		navigator:    navigator,
	}
}

type Core interface {
	Spawn(func()) bool
	Thread() *shredder.Thread
	CSPCollector() *common.CSPCollector
	ImportRegistry() *resources.Registry
	SessionId() string
	Conf() *common.SystemConf
	Id() string
	NewId() uint64
	Cinema() *door.Cinema
	NewLink(any) (*Link, error)
	SessionExpire(d time.Duration)
	SessionEnd()
	Call(call action.Call)
	SimpleCall(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc
	End()
	IsDetached() bool
}

type core[M any] struct {
	instance     coreInstance[M]
	store        *ctxstore.Store
	gen          *common.Primea
	hooksMu      sync.Mutex
	hooks        map[uint64]map[uint64]*door.DoorHook
	root         *door.Root
	solitaire    *solitaire
	navigator    *navigator[M]
	spawner      *shredder.Spawner
	cspCollector *common.CSPCollector
}

func (c *core[M]) Spawn(f func()) bool {
	return c.spawner.Go(f)
}
func (c *core[M]) IsDetached() bool {
	return c.navigator.isDetached()
}

func (c *core[M]) OnPanic(err error) {
	c.instance.OnPanic(err)
}

func (c *core[M]) SleepTimeout() time.Duration {
	return c.instance.conf().DisconnectHiddenTimer
}

func (c *core[M]) Conf() *common.SystemConf {
	return c.instance.conf()
}

func (c *core[M]) CSPCollector() *common.CSPCollector {
	return c.cspCollector
}

func (c *core[M]) ImportRegistry() *resources.Registry {
	return c.instance.getSession().getRouter().ImportRegistry()
}

func (c *core[M]) SessionId() string {
	return c.instance.getSession().Id()
}

func (c *core[M]) Cinema() *door.Cinema {
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

func (c *core[M]) NewLink(m any) (*Link, error) {
	return c.navigator.newLink(m)
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

func (c *core[M]) Call(call action.Call) {
	c.solitaire.Call(call)
}

func (c *core[M]) serve(w http.ResponseWriter, r *http.Request, page Page[M]) error {
	ctx := context.WithValue(context.Background(), common.CtxKeyInstance, c)
	ctx = c.store.Inject(ctx)
	ctx = c.instance.getSession().getStorage().Inject(ctx)
	ctx = context.WithValue(ctx, common.CtxKeyAdapters, c.instance.getSession().getRouter().Adapters())
	c.root = door.NewRoot(ctx, c)
	c.navigator.init(c.root.Ctx(), c)
	ch := c.root.Render(page.Render(c.navigator.getBeam()))
	render, ok := <-ch
	if !ok {
		return errors.New("instance killed before render")
	}
	render.InitImportMap(c.cspCollector)
	if c.cspCollector != nil {
		header := c.cspCollector.Generate()
		w.Header().Add("Content-Security-Policy", header)
		c.cspCollector = nil
	}
	gz := !c.instance.conf().ServerDisableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	if gz {
		w.Header().Set("Content-Encoding", "gzip")
	}
	if render.Err() != nil {
		return render.Err()
	}
	if code, ok := c.store.Load(common.CtxStorageKeyStatus).(int); ok {
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
	return nil
}
