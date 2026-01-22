package instance2

import (
	"log/slog"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
)

type killTimer struct {
	mu      sync.Mutex
	initial time.Duration
	regular time.Duration
	timer   *time.Timer
	inst    AnyInstance
}

func (t *killTimer) keepAlive() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.timer != nil {
		stopped := t.timer.Stop()
		if !stopped {
			return 
		}
		t.timer.Reset(t.regular)
		return 
	}
	t.timer = time.AfterFunc(t.initial, func() {
		slog.Debug("Inactive instance killed by timeout", slog.String("type", "message"), slog.String("instance_id", t.inst.Id()))
		t.inst.end(common.EndCauseKilled)
	})
}

