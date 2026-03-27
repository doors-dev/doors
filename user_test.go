package doors

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type helperDoor struct{}

func (helperDoor) Cinema() beam.Cinema {
	return nil
}

func (helperDoor) RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool, func(context.Context)) (core.Hook, bool) {
	return core.Hook{}, false
}

func (helperDoor) ID() uint64 {
	return 1
}

type helperInstance struct {
	expire   time.Duration
	adapters path.Adapters
	conf     common.SystemConf
}

func (h *helperInstance) CallCtx(context.Context, action.Action, func(json.RawMessage, error), func(), action.CallParams) context.CancelFunc {
	return func() {}
}

func (h *helperInstance) CallCheck(func() bool, action.Action, func(json.RawMessage, error), func(), action.CallParams) {
}

func (h *helperInstance) CSPCollector() *common.CSPCollector {
	return nil
}

func (h *helperInstance) ModuleRegistry() core.ModuleRegistry {
	return nil
}

func (h *helperInstance) ResourceRegistry() *resources.Registry {
	return nil
}

func (h *helperInstance) ID() string {
	return "instance-1"
}

func (h *helperInstance) RootID() uint64 {
	return 0
}

func (h *helperInstance) Conf() *common.SystemConf {
	return &h.conf
}

func (h *helperInstance) NewID() uint64 {
	return 2
}

func (h *helperInstance) NewLink(any) (core.Link, error) {
	return core.Link{}, nil
}

func (h *helperInstance) Runtime() shredder.Runtime {
	return nil
}

func (h *helperInstance) License() license.License {
	return nil
}

func (h *helperInstance) SetStatus(int) {}

func (h *helperInstance) SessionExpire(d time.Duration) {
	h.expire = d
}

func (h *helperInstance) SessionEnd() {}

func (h *helperInstance) InstanceEnd() {}

func (h *helperInstance) SessionID() string {
	return "session-1"
}

func (h *helperInstance) Adapters() path.Adapters {
	return h.adapters
}

func (h *helperInstance) PathMaker() path.PathMaker {
	return path.NewPathMaker("")
}

func (h *helperInstance) UpdateTitle(string, gox.Attrs) {}

func (h *helperInstance) UpdateMeta(string, bool, gox.Attrs) {}

type helperLocation struct {
	Home bool   `path:"/home"`
	Tag  string `query:"tag"`
}

func helperContext(t *testing.T, adapters path.Adapters) (context.Context, *helperInstance) {
	t.Helper()
	inst := &helperInstance{adapters: adapters}
	return context.WithValue(context.Background(), ctex.KeyCore, core.NewCore(inst, helperDoor{})), inst
}

func TestUserHelpers(t *testing.T) {
	adapter, err := path.NewAdapter[helperLocation]()
	if err != nil {
		t.Fatal(err)
	}
	var adapters path.Adapters
	adapters.Add(adapter)

	ctx, inst := helperContext(t, adapters)
	SessionExpire(ctx, time.Hour)
	if inst.expire != time.Hour {
		t.Fatalf("unexpected session expire duration: %v", inst.expire)
	}

	location, err := NewLocation(ctx, helperLocation{Home: true, Tag: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if location.String() != "/home?tag=x" {
		t.Fatalf("unexpected location: %q", location.String())
	}

	emptyCtx, _ := helperContext(t, nil)
	if _, err := NewLocation(emptyCtx, helperLocation{Home: true}); err == nil {
		t.Fatal("expected missing adapter to fail")
	}

	if IDRand() == "" {
		t.Fatal("expected random id to be non-empty")
	}
	if IDRand() == IDRand() {
		t.Fatal("expected random ids to differ")
	}
	if IDString("hello") != IDString("hello") {
		t.Fatal("expected deterministic string id")
	}
	if IDString("hello") == IDString("world") {
		t.Fatal("expected distinct string ids")
	}
	if IDBytes([]byte("hello")) != IDBytes([]byte("hello")) {
		t.Fatal("expected deterministic byte id")
	}
	if IDBytes([]byte("hello")) == IDBytes([]byte("world")) {
		t.Fatal("expected distinct byte ids")
	}

	if ctex.IsBlockingCtx(context.Background()) {
		t.Fatal("unexpected blocking context by default")
	}
	if !ctex.IsBlockingCtx(AllowBlocking(context.Background())) {
		t.Fatal("expected AllowBlocking to mark context as blocking")
	}
}
