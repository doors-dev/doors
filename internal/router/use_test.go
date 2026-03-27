package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/router/model"
	"github.com/evanw/esbuild/pkg/api"
)

type stubSessionCallback struct {
	deleted []string
}

func (s *stubSessionCallback) Create(string, http.Header) {}

func (s *stubSessionCallback) Delete(id string) {
	s.deleted = append(s.deleted, id)
}

type stubProfiles struct{}

func (stubProfiles) Options(string) api.BuildOptions {
	return api.BuildOptions{
		Target: api.ES2022,
	}
}

type stubRoute struct {
	called bool
}

func (s *stubRoute) Match(r *http.Request) bool {
	return r.URL.Path == "/route"
}

func (s *stubRoute) Serve(w http.ResponseWriter, r *http.Request) {
	s.called = true
	w.WriteHeader(http.StatusTeapot)
}

func TestUseHelpersAndRouterState(t *testing.T) {
	sessionHooks{}.Create("id", http.Header{})
	sessionHooks{}.Delete("id")

	router := NewRouter()

	fallbackCalled := false
	router.Use(UseFallback(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalled = true
		w.WriteHeader(http.StatusAccepted)
	})))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	router.ServeHTTP(recorder, request)
	if !fallbackCalled {
		t.Fatal("expected fallback handler to be used")
	}
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unexpected fallback status: %d", recorder.Code)
	}

	router.Use(UseSystemConf(common.SystemConf{RequestTimeout: -1}))
	if router.Conf().RequestTimeout != 30*time.Second {
		t.Fatalf("expected system defaults to be applied, got %s", router.Conf().RequestTimeout)
	}
	if router.Conf().SessionInstanceLimit == 0 {
		t.Fatal("expected system defaults to initialize session limit")
	}

	sessionCallback := &stubSessionCallback{}
	router.Use(UseSessionCallback(sessionCallback))
	router.sessions.Store("session-1", struct{}{})
	router.RemoveSession("session-1")
	if len(sessionCallback.deleted) != 1 || sessionCallback.deleted[0] != "session-1" {
		t.Fatalf("unexpected deleted session ids: %#v", sessionCallback.deleted)
	}
	if _, ok := router.sessions.Load("session-1"); ok {
		t.Fatal("expected session to be removed")
	}

	router.Use(UseESConf(stubProfiles{}))
	if _, ok := router.BuildProfiles().(stubProfiles); !ok {
		t.Fatal("expected custom build profiles to be stored")
	}

	csp := &common.CSP{}
	router.Use(UseCSP(csp))
	if router.CSP() != csp {
		t.Fatal("expected CSP config to be stored")
	}

	router.Use(UseServerID("server-1"))
	if got := router.PathMaker().ID(); got != "server-1" {
		t.Fatalf("unexpected server id: %q", got)
	}
}

func TestUseModelAndRoute(t *testing.T) {
	type page struct {
		Home bool `path:""`
	}

	router := NewRouter()
	router.Use(UseModel(func(w http.ResponseWriter, r *http.Request, source beam.Source[page], store ctex.Store) model.Res {
		return model.Res{}
	}))
	if len(router.modelRoutes) != 1 {
		t.Fatalf("expected one model route, got %d", len(router.modelRoutes))
	}
	if len(router.Adapters()) != 1 {
		t.Fatalf("expected one model adapter, got %d", len(router.Adapters()))
	}

	route := &stubRoute{}
	router.Use(UseRoute(route))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/route", nil)
	if !router.tryServeRoute(recorder, request) {
		t.Fatal("expected custom route to match")
	}
	if !route.called {
		t.Fatal("expected route handler to be called")
	}
	if recorder.Code != http.StatusTeapot {
		t.Fatalf("unexpected route status: %d", recorder.Code)
	}
}

func TestUseLicenseAndServerIDValidation(t *testing.T) {
	router := NewRouter()

	UseLicense("Doors Commercial").apply(router)
	if router.License() != "Doors Commercial" {
		t.Fatal("expected license string to be stored")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected invalid server id to panic")
		}
	}()
	UseServerID("bad/id")
}

var _ resources.BuildProfiles = stubProfiles{}
