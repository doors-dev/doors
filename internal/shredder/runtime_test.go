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
	"testing"
)

func TestNilRuntimeSubmitSpawns(t *testing.T) {
	done := make(chan bool, 1)
	FreeFrame{}.Submit(context.Background(), nil, func(ok bool) {
		done <- ok
	})
	if !<-done {
		t.Fatal("expected task to run with ok=true")
	}
}

func TestNilRuntimeSubmitCanceledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan bool, 1)
	FreeFrame{}.Submit(ctx, nil, func(ok bool) {
		done <- ok
	})
	if <-done {
		t.Fatal("expected task to run with ok=false on canceled ctx")
	}
}

func TestNilRuntimeSubmitPanicRecovered(t *testing.T) {
	done := make(chan struct{})
	FreeFrame{}.Submit(context.Background(), nil, func(bool) {
		defer close(done)
		panic("boom")
	})
	<-done
}
