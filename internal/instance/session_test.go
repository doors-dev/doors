package instance

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
)

type sessionTestApp struct {
	conf    common.Conf
	removed chan string
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
	return path.NewPathMaker("")
}

func (a *sessionTestApp) RemoveSession(id string) {
	a.removed <- id
}

func (a *sessionTestApp) ResourceRegistry() resources.Registry {
	return nil
}

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

type noopResponseWriter struct{}

func (noopResponseWriter) Header() http.Header {
	return http.Header{}
}

func (noopResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (noopResponseWriter) WriteHeader(int) {}
