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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

type syncLog struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (l *syncLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

func (l *syncLog) lines(msg string) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []string
	for line := range strings.SplitSeq(l.buf.String(), "\n") {
		if strings.Contains(line, msg) {
			out = append(out, line)
		}
	}
	return out
}

type emptyComp struct{}

func (emptyComp) Main() gox.Elem {
	return nil
}

type stubCall struct{}

func (stubCall) Params() actions.CallParams {
	return actions.CallParams{}
}

func (stubCall) Action() (actions.Action, func(), bool) {
	return nil, func() {}, false
}

func (stubCall) Cancel() {}

func (stubCall) Result(json.RawMessage, error) {}

func servedInstance(t *testing.T, queue int) (Instance, *syncLog) {
	t.Helper()
	log := &syncLog{}
	app := newSessionTestApp()
	app.conf.SolitaireQueue = queue
	app.logger = slog.New(slog.NewTextHandler(log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	sess := NewSession(app, "test-session")
	inst, ok := sess.Instance(path.Location{}, false)
	if !ok {
		t.Fatal("expected an instance")
	}
	page := func(context.Context, http.ResponseWriter, *http.Request) gox.Comp {
		return emptyComp{}
	}
	err, handled := inst.Serve(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), page)
	if err != nil || !handled {
		t.Fatalf("serve failed: err=%v handled=%v", err, handled)
	}
	return inst, log
}

const syncErrorMsg = "Instance synchronization error"

func TestSyncErrorLogsOnlyTheEndingError(t *testing.T) {
	inst, log := servedInstance(t, 1024)
	for range 3 {
		inst.SyncError(errors.New("sync timeout"))
	}
	if lines := log.lines(syncErrorMsg); len(lines) != 1 {
		t.Fatalf("expected one sync error line, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}
}

func TestCallsAfterKillLogNoSyncError(t *testing.T) {
	inst, log := servedInstance(t, 1024)
	inst.Kill()
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			inst.Call(stubCall{}, 0, 0, 0, 0, 0)
		})
	}
	wg.Wait()
	if lines := log.lines(syncErrorMsg); len(lines) != 0 {
		t.Fatalf("expected no sync error lines, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}
}

func TestQueueOverflowLogsOnceWithHint(t *testing.T) {
	inst, log := servedInstance(t, 4)
	for range 10 {
		inst.Call(stubCall{}, 0, 0, 0, 0, 0)
	}
	lines := log.lines(syncErrorMsg)
	if len(lines) != 1 {
		t.Fatalf("expected one sync error line, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[0], "call queue limit reached") || !strings.Contains(lines[0], "OnBatch") {
		t.Fatalf("expected the queue limit hint, got %s", lines[0])
	}
}
