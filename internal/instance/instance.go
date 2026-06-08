package instance

import (
	"compress/gzip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/instance/utils"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/doors/internal/solitaire"
	"github.com/doors-dev/gox"
)

type Instance = *instance

func newInstance(sess *session, loc path.Location) Instance {
	return &instance{
		id:       common.RandId(),
		session:  sess,
		store:    ctex.NewStore(),
		location: beam.NewSource(loc, path.EqualLocation, false),
		prime:    common.NewPrime(),
	}
}

const (
	zero int32 = iota
	initializing
	active
	killed
)

type instance struct {
	state      atomic.Int32
	id         string
	session    *session
	store      ctex.Store
	location   beam.Source[path.Location]
	prime      common.Prime
	runtime    shredder.Runtime
	solitaire  solitaire.Solitaire
	root       door.Root
	navigator  utils.Navigator
	killTimer  utils.KillTimer
	csp        common.CSPCollector
	importMap  utils.ImportMap
	pageStatus atomic.Int32
	titleMeta  core.TitleMeta
}

func (inst Instance) Logger() *slog.Logger {
	return inst.session.app.Logger()
}

func (inst Instance) UpdateLocation(l path.Location) bool {
	return inst.navigator.Update(l)
}

func (inst Instance) TriggerHook(hookID uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	if inst.state.Load() != active {
		return false
	}
	ok := inst.root.TriggerHook(hookID, w, r, track)
	if ok {
		inst.session.limiter.TouchHeavy(inst.id)
	}
	return ok

}

func (inst Instance) Connect(w http.ResponseWriter, r *http.Request) {
	if inst.state.Load() != active {
		w.WriteHeader(http.StatusGone)
		return
	}
	inst.killTimer.KeepAlive()
	inst.solitaire.Connect(w, r)
}

func (inst Instance) CSPCollector() common.CSPCollector {
	return inst.csp
}

func (inst *instance) Call(call action.Call) {
	inst.solitaire.Call(call)
}

func (inst *instance) Kill() {
	inst.end(common.EndCauseKilled)
}

func (inst *instance) ModuleRegistry() core.ModuleRegistry {
	return inst.importMap
}

func (inst *instance) NewID() uint64 {
	return inst.prime.Gen()
}

func (inst *instance) RootID() uint64 {
	return inst.root.ID()
}

func (inst *instance) Runtime() shredder.Runtime {
	return inst.runtime
}

func (inst *instance) SetStatus(s int) {
	inst.pageStatus.Store(int32(s))
}

func (inst *instance) SyncError(err error) {
	inst.Logger().Debug("Instance synchronization error", "error", err, "type", "error", "instance_id", inst.id)
	inst.end(common.EndCauseSyncError)
}

func (inst *instance) Touch() {
	inst.session.limiter.TouchLight(inst.id)
}

func (inst Instance) Session() core.Session {
	return inst.session
}

func (inst Instance) Location() beam.Source[path.Location] {
	return inst.location
}

func (inst Instance) ID() string {
	return inst.id
}

func (inst Instance) Store() ctex.Store {
	return inst.store
}

func (inst Instance) TitleMeta() core.TitleMeta {
	return inst.titleMeta
}

type Page = func(ctx context.Context, w http.ResponseWriter, r *http.Request) gox.Comp

type instanceComp struct {
	w    http.ResponseWriter
	r    *http.Request
	page Page
	inst *instance
}

func (i instanceComp) Main() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		ctx := cur.Context()
		i.inst.navigator = utils.NewNavigator(i.inst, ctx)
		comp := i.page(ctx, i.w, i.r)
		i.inst.navigator.Sync()
		el := comp.Main()
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

func (inst Instance) Serve(w http.ResponseWriter, r *http.Request, page Page) (err error, handeled bool) {
	if !inst.state.CompareAndSwap(zero, initializing) {
		return nil, false
	}
	inst.session.app.InstanceCreated()
	inst.runtime = shredder.NewRuntime(inst.session.Context(), inst.Session().App().Conf().InstanceGoroutineLimit, inst)
	inst.solitaire = solitaire.NewSolitaire(inst, common.GetSolitaireConf(inst.Session().App().Conf()))
	inst.root = door.NewRoot(inst)
	inst.killTimer = utils.NewKillTimer(inst)
	inst.csp = inst.session.app.CSP().NewCollector()
	inst.importMap = utils.NewImportMap()
	inst.titleMeta = utils.NewTitleMeta(inst)
	stack, err := inst.root.Render(r.Context(), instanceComp{
		w:    w,
		r:    r,
		page: page,
		inst: inst,
	})
	if err != nil {
		ok := inst.state.CompareAndSwap(initializing, killed)
		inst.clean(common.EndCauseKilled)
		return err, ok
	}
	if !inst.state.CompareAndSwap(initializing, active) {
		inst.clean(common.EndCauseSuspend)
		return nil, false
	}
	static := inst.root.IsStatic()
	if !static {
		inst.killTimer.KeepAlive()
	}
	if err := inst.render(w, r, stack, static); err != nil {
		inst.end(common.EndCauseKilled)
		return nil, true
	}
	if static {
		inst.end(common.EndCauseKilled)
		return nil, true
	}
	return nil, true
}

func (inst *instance) render(w http.ResponseWriter, r *http.Request, pipe door.Stack, static bool) error {
	gz := !inst.session.App().Conf().ServerDisableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	importMap, importHash := inst.importMap.Generate()
	inst.renderHeaders(w, gz, importHash)
	var writer io.Writer = w
	if gz {
		wgz := gzip.NewWriter(w)
		defer wgz.Close()
		writer = wgz
	}
	pr := printer.NewPagePrinter(writer, static, front.Include(inst), importMap, inst.titleMeta)
	return pipe.Print(pr)
}

func (inst *instance) renderHeaders(w http.ResponseWriter, gz bool, importHash []byte) {
	if inst.csp != nil {
		if importHash != nil {
			inst.csp.ScriptHash(importHash)
		}
		header := inst.csp.Generate()
		w.Header().Add("Content-Security-Policy", header)
		inst.csp = nil
	}
	if gz {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.WriteHeader(inst.getStatus())
}

func (inst *instance) getStatus() int {
	if s := inst.pageStatus.Load(); s > 0 {
		return int(s)
	}
	return http.StatusOK
}

func (inst *instance) end(cause common.EndCause) {
	prev := inst.state.Swap(killed)
	if prev != active {
		return
	}
	inst.clean(cause)
}

func (inst Instance) clean(cause common.EndCause) {
	inst.session.removeInstance(inst.id)
	inst.runtime.Cancel()
	inst.solitaire.End(cause)
	inst.killTimer.Stop()
	inst.root.Kill()
	inst.session.app.InstanceDeleted()
}
