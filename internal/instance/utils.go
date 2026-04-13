// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		slog.Debug("inactive instance killed by timeout", "type", "message", "instance_id", t.inst.ID())
		t.inst.end(common.EndCauseKilled)
	})
}

func newImportMap() *importMap {
	return &importMap{
		Imports: make(map[string]string),
	}
}

type importMap struct {
	mu      sync.Mutex        `json:"-"`
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
