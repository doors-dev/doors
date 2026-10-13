// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package batch

import (
	"bytes"
	"context"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	dpath "github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

const retryCloser = "report_hook"

type retryLoss struct {
	mu        sync.Mutex
	marker    []byte
	hold      chan struct{}
	dropped   chan struct{}
	delivered [][]byte
}

func (l *retryLoss) arm(marker string, hold bool) chan struct{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.marker = []byte(marker)
	l.hold = nil
	if hold {
		l.hold = make(chan struct{})
	}
	return l.hold
}

func (l *retryLoss) drop(p []byte) bool {
	l.mu.Lock()
	if l.marker == nil || !bytes.Contains(p, l.marker) {
		l.delivered = append(l.delivered, bytes.Clone(p))
		l.mu.Unlock()
		return false
	}
	l.marker = nil
	hold := l.hold
	l.mu.Unlock()
	l.dropped <- struct{}{}
	if hold != nil {
		select {
		case <-hold:
		case <-time.After(2 * waitLimit):
		}
	}
	return true
}

func (l *retryLoss) count(marker string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for _, p := range l.delivered {
		n += bytes.Count(p, []byte(marker))
	}
	return n
}

func (l *retryLoss) waitDropped(t *testing.T) {
	t.Helper()
	select {
	case <-l.dropped:
	case <-time.After(waitLimit):
		t.Fatal("no frame was dropped")
	}
}

func (l *retryLoss) waitDelivered(t *testing.T, marker string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for l.count(marker) == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("%q was not delivered", marker)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

type retryWriter struct {
	http.ResponseWriter
	l *retryLoss
}

func (w retryWriter) Write(p []byte) (int, error) {
	if w.l.drop(p) {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}

func (w retryWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w retryWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func retryOpen(t *testing.T) (*rod.Page, *retryState, *retryLoss) {
	t.Helper()
	s := retryNewState()
	l := &retryLoss{dropped: make(chan struct{}, 8)}
	adapter, err := dpath.GetModelAdapter[test.Path]()
	if err != nil {
		t.Fatal(err)
	}
	bro := test.NewBroWrapOptions(browser, func(ctx context.Context, _ doors.Request) gox.Comp {
		lens := doors.DeriveSourceEqual(doors.Router(ctx), func(loc doors.Location) test.Path {
			m, ok := adapter.Decode(loc)
			if !ok {
				return test.Path{}
			}
			return *m
		}, func(prev doors.Location, m test.Path) doors.Location {
			loc, err := adapter.Encode(&m)
			if err != nil {
				return prev
			}
			return loc
		}, func(a test.Path, b test.Path) bool {
			return reflect.DeepEqual(a, b)
		})
		return &test.Page{
			Source: lens,
			F:      &retryFragment{s: s},
		}
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/s/") {
				w = retryWriter{ResponseWriter: w, l: l}
			}
			next.ServeHTTP(w, r)
		})
	}, []doors.With{doors.WithConf(doors.Conf{SolitaireDisableGzip: true})})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	limitsWatch(t, page, "#x", "#y")
	return page, s, l
}

func retryHoldBatch(t *testing.T, page *rod.Page, s *retryState, l *retryLoss) {
	t.Helper()
	test.ClickNow(t, page, "#retry-batch")
	limitsEntered(t, s.entered, "the batch handler")
	l.waitDelivered(t, retryBatchX)
	time.Sleep(quiet)
	limitsHeld(t, page, "#x", "x0")
}

func retryApplied(t *testing.T, page *rod.Page, x string) {
	t.Helper()
	waitContent(t, page, "#y", retryBatchY)
	waitContent(t, page, "#x", x)
	time.Sleep(quiet)
	if got := limitsSeen(t, page); got != "x0|y0 "+x+"|"+retryBatchY {
		t.Fatalf("expected one joint apply ending on %q, saw %q", x, got)
	}
}

// An update lost behind an open batch and superseded before its resend is replaced by a filler, and the page ends on the superseding update.
func TestRetrySupersededLostUpdate(t *testing.T) {
	page, s, l := retryOpen(t)
	retryHoldBatch(t, page, s, l)
	hold := l.arm(retrySecondX, true)
	test.ClickNow(t, page, "#retry-second")
	l.waitDropped(t)
	test.ClickNow(t, page, "#retry-third")
	limitsEntered(t, s.third, "the third update")
	close(hold)
	l.waitDelivered(t, retryThirdX)
	time.Sleep(quiet)
	limitsHeld(t, page, "#x", "x0")
	if got := l.count(retrySecondX); got != 0 {
		t.Fatalf("expected the superseded update not sent again, sent %d times", got)
	}
	limitsProceed(t, s.proceed, "the batch handler")
	retryApplied(t, page, retryThirdX)
}

// An update of a held door whose frame is lost is sent again, still waits for the batch and lands after it.
func TestRetryLostUpdateBehindBatch(t *testing.T) {
	page, s, l := retryOpen(t)
	retryHoldBatch(t, page, s, l)
	l.arm(retrySecondX, false)
	test.ClickNow(t, page, "#retry-second")
	l.waitDropped(t)
	l.waitDelivered(t, retrySecondX)
	time.Sleep(quiet)
	limitsHeld(t, page, "#x", "x0")
	limitsProceed(t, s.proceed, "the batch handler")
	retryApplied(t, page, retrySecondX)
}

// A lost frame with batch members is sent again while the batch is open, and the batch applies once, together.
func TestRetryLostMemberRecovered(t *testing.T) {
	page, s, l := retryOpen(t)
	l.arm(retryBatchX, false)
	test.ClickNow(t, page, "#retry-batch")
	l.waitDropped(t)
	limitsEntered(t, s.entered, "the batch handler")
	l.waitDelivered(t, retryBatchX)
	time.Sleep(quiet)
	limitsHeld(t, page, "#x", "x0")
	limitsProceed(t, s.proceed, "the batch handler")
	retryApplied(t, page, retryBatchX)
	if got := l.count(retryBatchX); got != 1 {
		t.Fatalf("expected the lost member delivered once, delivered %d times", got)
	}
}

// Batch members lost and recovered only after the closer was issued still apply with the batch, once.
func TestRetryLostMemberBeforeCloser(t *testing.T) {
	page, s, l := retryOpen(t)
	hold := l.arm(retryBatchX, true)
	test.ClickNow(t, page, "#retry-batch")
	l.waitDropped(t)
	limitsEntered(t, s.entered, "the batch handler")
	limitsProceed(t, s.proceed, "the batch handler")
	time.Sleep(quiet)
	limitsHeld(t, page, "#x", "x0")
	limitsHeld(t, page, "#y", "y0")
	close(hold)
	retryApplied(t, page, retryBatchX)
	if got := l.count(retryBatchX); got != 1 {
		t.Fatalf("expected the lost member delivered once, delivered %d times", got)
	}
}

// A lost closer is sent again and the batch applies once, together.
func TestRetryLostCloserRecovered(t *testing.T) {
	page, s, l := retryOpen(t)
	retryHoldBatch(t, page, s, l)
	l.arm(retryCloser, false)
	limitsProceed(t, s.proceed, "the batch handler")
	l.waitDropped(t)
	if got := l.count(retryCloser); got != 0 {
		t.Fatalf("expected the closer lost before any delivery, delivered %d times", got)
	}
	retryApplied(t, page, retryBatchX)
	if got := l.count(retryCloser); got != 1 {
		t.Fatalf("expected the closer delivered once, delivered %d times", got)
	}
}
