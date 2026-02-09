// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"crypto/sha256"
	"encoding/json"
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
		slog.Debug("Inactive instance killed by timeout", slog.String("type", "message"), slog.String("instance_id", t.inst.ID()))
		t.inst.end(common.EndCauseKilled)
	})
}


func newImportMap() *importMap {
	return &importMap{
		Imports: make(map[string]string),
	}
}

type importMap struct{
	mu sync.Mutex `json:"-"`
	Imports map[string]string `json:"imports"`
}


func (i *importMap) Add(specifier string, path string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Imports[specifier] = path
}


func (i *importMap) generate() (content []byte, hash []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if len(i.Imports) == 0 {
		return
	}
	content, _ = json.Marshal(i)
	sum := sha256.Sum256(content)
	return content, sum[:]
}

