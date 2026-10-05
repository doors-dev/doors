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

package shredder

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type warningCounter struct {
	n *atomic.Int64
}

func (h warningCounter) Enabled(context.Context, slog.Level) bool { return true }

func (h warningCounter) Handle(_ context.Context, r slog.Record) error {
	if r.Level >= slog.LevelWarn {
		h.n.Add(1)
	}
	return nil
}

func (h warningCounter) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h warningCounter) WithGroup(string) slog.Handler { return h }

type threadStress struct {
	t          *testing.T
	warnings   atomic.Int64
	writing    atomic.Int32
	reading    atomic.Int32
	violations atomic.Int64
	stop       chan struct{}
	readers    sync.WaitGroup
}

func newThreadStress(t *testing.T) *threadStress {
	s := &threadStress{
		t:    t,
		stop: make(chan struct{}),
	}
	prev := slog.Default()
	slog.SetDefault(slog.New(warningCounter{n: &s.warnings}))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return s
}

func (s *threadStress) run(f ReleaseFrame, step func()) bool {
	done := make(chan struct{})
	f.Run(nil, nil, func(bool) {
		step()
		close(done)
	})
	f.Release()
	select {
	case <-done:
		return true
	case <-time.After(5 * time.Second):
		s.t.Error("frame did not run")
		return false
	}
}

func (s *threadStress) readStep() {
	s.reading.Add(1)
	if s.writing.Load() != 0 {
		s.violations.Add(1)
	}
	s.reading.Add(-1)
}

func (s *threadStress) writeStep() {
	if s.writing.Add(1) != 1 || s.reading.Load() != 0 {
		s.violations.Add(1)
	}
	s.writing.Add(-1)
}

func (s *threadStress) startReaders(n int, read func() ReleaseFrame) {
	for i := range n {
		s.readers.Add(1)
		go func() {
			defer s.readers.Done()
			for {
				select {
				case <-s.stop:
					return
				default:
				}
				if !s.run(read(), s.readStep) {
					return
				}
				if i%2 == 0 {
					time.Sleep(time.Duration(i) * time.Microsecond)
				}
			}
		}()
	}
}

func (s *threadStress) concurrentWriters(n int, writes int, write func() ReleaseFrame) {
	var writers sync.WaitGroup
	for range n {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for range writes {
				if !s.run(write(), s.writeStep) {
					return
				}
			}
		}()
	}
	writers.Wait()
}

func (s *threadStress) finish() {
	close(s.stop)
	s.readers.Wait()
	if v := s.violations.Load(); v != 0 {
		s.t.Fatalf("reads and writes overlapped %d times", v)
	}
	if w := s.warnings.Load(); w != 0 {
		s.t.Fatalf("scheduled on completed frames %d times", w)
	}
}

func TestReadBlockingWriteThreadStress(t *testing.T) {
	var thread ReadBlockingWriteThread
	s := newThreadStress(t)
	s.startReaders(6, thread.Read)
	for range 300 {
		write, read := thread.Write()
		write.Run(nil, nil, func(bool) { s.writeStep() })
		write.Release()
		if !s.run(read, func() {}) {
			break
		}
	}
	s.finish()
}

func TestReadStarveWriteThreadStress(t *testing.T) {
	var thread ReadStarveWriteThread
	s := newThreadStress(t)
	s.startReaders(6, thread.Read)
	s.concurrentWriters(2, 300, thread.Write)
	s.finish()
}

func TestReadWriteThreadStress(t *testing.T) {
	var thread ReadWriteThread
	s := newThreadStress(t)
	s.startReaders(6, thread.Read)
	s.concurrentWriters(2, 300, thread.Write)
	s.finish()
}
