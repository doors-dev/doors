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
)

type sessionTestApp struct {
	conf       common.Conf
	cookieName string
	removed    chan string
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

func TestSessionKillCancelsContext(t *testing.T) {
	app := newSessionTestApp()
	sess := NewSession(app)
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

func TestSessionRenewCookie(t *testing.T) {
	app := newSessionTestApp()
	sess := NewSession(app)
	w := httptest.NewRecorder()
	if !sess.Renew(w) {
		t.Fatal("expected live session to renew")
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one session cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "test" {
		t.Fatalf("unexpected cookie name: %q", cookie.Name)
	}
	if cookie.Value != sess.ID() {
		t.Fatalf("unexpected cookie value: %q", cookie.Value)
	}
	if !cookie.Secure {
		t.Fatal("expected session cookie to be secure")
	}
	if cookie.Path != "/" {
		t.Fatalf("unexpected cookie path: %q", cookie.Path)
	}

	app.conf.ServerSessionCookieNoSecure = true
	app.conf.ServerSessionCookiePrefix = ""
	w = httptest.NewRecorder()
	sess = NewSession(app)
	if !sess.Renew(w) {
		t.Fatal("expected live no-secure session to renew")
	}
	cookies = w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one no-secure session cookie, got %d", len(cookies))
	}
	cookie = cookies[0]
	if cookie.Name != "test" {
		t.Fatalf("unexpected no-secure cookie name: %q", cookie.Name)
	}
	if cookie.Secure {
		t.Fatal("expected no-secure session cookie to omit Secure")
	}
}

func TestSessionRenewServerIDCookie(t *testing.T) {
	app := newSessionTestApp()
	app.cookieName = "server_id"
	sess := NewSession(app)
	w := httptest.NewRecorder()
	if !sess.Renew(w) {
		t.Fatal("expected live session to renew")
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected two cookies (session + server ID), got %d", len(cookies))
	}
	serverCookie := cookies[1]
	if serverCookie.Name != "server_id" {
		t.Fatalf("unexpected server ID cookie name: %q", serverCookie.Name)
	}
	if serverCookie.Value != "test" {
		t.Fatalf("unexpected server ID cookie value: %q", serverCookie.Value)
	}
	if !serverCookie.HttpOnly {
		t.Fatal("expected server ID cookie to be HttpOnly")
	}
	if !serverCookie.Secure {
		t.Fatal("expected server ID cookie to be Secure")
	}
	if serverCookie.Path != "/" {
		t.Fatalf("unexpected server ID cookie path: %q", serverCookie.Path)
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
