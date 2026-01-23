package instance2

import (
	"crypto/sha256"
	"encoding/json"
	"io"
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


func newImportMap() *moduleImportMap {
	return &moduleImportMap{
		storage: make(map[string]string),
	}
}

type moduleImportMap struct{
	mu sync.Mutex
	storage map[string]string
}


func (i *moduleImportMap) add(specifier string, path string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.storage[specifier] = path
}


func (i *moduleImportMap) generate() (content []byte, hash []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if len(i.storage) == 0 {
		return
	}
	content, _ = json.Marshal(i)
	sum := sha256.Sum256(content)
	return content, sum[:]
}

