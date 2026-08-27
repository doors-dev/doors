package instance

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type sessionTestApp struct {
	conf         common.Conf
	cookieName   string
	removed      chan string
	cookieID     string
	cookieMaxAge time.Duration
}

func newSessionTestApp() *sessionTestApp {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	return &sessionTestApp{
		conf:    conf,
		removed: make(chan string, 1),
	}
}

func (a *sessionTestApp) CSP() *common.CSP {
	return &common.CSP{}
}

func (a *sessionTestApp) Conf() *common.Conf {
	return &a.conf
}

func (a *sessionTestApp) PathMaker() path.PathMaker {
	return path.NewPathMaker(a.conf.ServerSessionCookiePrefix, "test", a.cookieName)
}

func (a *sessionTestApp) RemoveSession(id string) {
	a.removed <- id
}

func (a *sessionTestApp) ResourceRegistry() resources.Registry {
	return nil
}

func (a *sessionTestApp) Logger() *slog.Logger {
	return slog.Default()
}

func (a *sessionTestApp) InstanceCreated() {}
func (a *sessionTestApp) InstanceDeleted() {}
func (a *sessionTestApp) Draining() bool   { return false }
func (a *sessionTestApp) SetCookies(w http.ResponseWriter, id string, maxAge time.Duration) {
	a.cookieID = id
	a.cookieMaxAge = maxAge
}

func (a *sessionTestApp) PrinterMiddleware() func(next gox.Printer) gox.Printer {
	return func(next gox.Printer) gox.Printer { return next }
}

func TestSessionKillCancelsContext(t *testing.T) {
	app := newSessionTestApp()
	sess := NewSession(app, "test-session")
	ctx := sess.Context()

	select {
	case <-ctx.Done():
		t.Fatal("expected live session context before Kill")
	default:
	}

	sess.Kill()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("expected Kill to cancel session context")
	}
	if ctx.Err() != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", ctx.Err())
	}

	select {
	case id := <-app.removed:
		if id != sess.ID() {
			t.Fatalf("expected removed session %q, got %q", sess.ID(), id)
		}
	case <-time.After(time.Second):
		t.Fatal("expected Kill to remove session from app")
	}

	if _, ok := sess.Instance(path.Location{}); ok {
		t.Fatal("expected killed session to reject new instances")
	}
	if sess.Renew(noopResponseWriter{}) {
		t.Fatal("expected killed session to reject renewal")
	}
}

func TestSessionRenewSetsCookies(t *testing.T) {
	app := newSessionTestApp()
	sess := NewSession(app, "test-session")
	sess.renewed.Store(0)
	w := httptest.NewRecorder()
	if !sess.Renew(w) {
		t.Fatal("expected live session to renew")
	}
	if app.cookieID != "test-session" {
		t.Fatalf("unexpected cookie id: %q", app.cookieID)
	}
	if app.cookieMaxAge <= 0 || app.cookieMaxAge > app.conf.SessionTTL {
		t.Fatalf("unexpected cookie max age: %v", app.cookieMaxAge)
	}
}

type noopResponseWriter struct{}

func (noopResponseWriter) Header() http.Header {
	return http.Header{}
}

func (noopResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (noopResponseWriter) WriteHeader(int) {}
