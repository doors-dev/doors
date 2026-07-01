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

package utils

import (
	"crypto/sha256"
	"encoding/json"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/core"
)

type KillTimer = *killTimer

func NewKillTimer(inst core.Instance) KillTimer {
	conf := inst.Session().App().Conf()
	t := &killTimer{
		initial: conf.InstanceConnectTimeout,
		regular: conf.InstanceTTL,
		inst:    inst,
	}
	return t
}

type killTimer struct {
	mu        sync.Mutex
	stopped   bool
	initial   time.Duration
	regular   time.Duration
	timer     *time.Timer
	inst      core.Instance
	keepAlive time.Time
}

func (t *killTimer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.stopped = true
	if t.timer == nil {
		return
	}
	t.timer.Stop()
}

func (t *killTimer) LastSeen() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.keepAlive
}

func (t *killTimer) KeepAlive() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.keepAlive = time.Now()
	if t.timer != nil {
		stopped := t.timer.Stop()
		if !stopped {
			return
		}
		t.timer.Reset(t.regular)
		return
	}
	t.timer = time.AfterFunc(t.initial, func() {
		t.inst.Logger().Debug("inactive instance killed by timeout", "type", "message", "instance_id", t.inst.ID())
		t.inst.Kill()
	})
}

type ImportMap = *importMap

func NewImportMap() *importMap {
	return &importMap{
		Imports: make(map[string]string),
	}
}

type importMap struct {
	mu      sync.Mutex        `json:"-"`
	Imports map[string]string `json:"imports"`
}

func (i ImportMap) Add(specifier string, path string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Imports[specifier] = path
}

func (i ImportMap) Generate() (content []byte, hash []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if len(i.Imports) == 0 {
		return
	}
	content, _ = json.Marshal(i)
	sum := sha256.Sum256(content)
	return content, sum[:]
}
