package app

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/instance"
)

func getSession(ctx context.Context) (instance.Session, bool) {
	switch s := ctx.Value(common.KeySession).(type) {
	case instance.Session:
		return s, true
	case *lazySession:
		return s.get(), true
	default:
		return nil, false
	}
}

type lazySession struct {
	id      string
	r       *http.Request
	app     *app
	once    sync.Once
	session instance.Session
}

var _ core.Session = (*lazySession)(nil)

func (l *lazySession) Context() context.Context {
	return l.get().Context()
}

func (l *lazySession) Expire(d time.Duration) {
	l.get().Expire(d)
}

func (l *lazySession) ID() string {
	return l.get().ID()
}

func (l *lazySession) Kill() {
	l.get().Kill()
}

func (l *lazySession) LastSeen() time.Time {
	return l.get().LastSeen()
}

func (l *lazySession) Store() ctex.Store {
	return l.get().Store()
}

func (l *lazySession) App() core.App {
	return l.app
}

func (l *lazySession) Logger() *slog.Logger {
	return l.app.logger
}

func (l *lazySession) get() instance.Session {
	l.once.Do(func() {
		l.session = l.app.newSession(l.r, l.id)
	})
	return l.session
}
