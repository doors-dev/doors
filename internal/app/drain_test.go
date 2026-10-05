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

package app

import (
	"log/slog"
	"testing"
	"time"
)

func newDrainTestApp() *app {
	return &app{logger: slog.Default()}
}

func waitDrainCallback(t *testing.T, called <-chan struct{}) {
	t.Helper()
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("drain callback not called")
	}
}

func TestDrainNatural(t *testing.T) {
	a := newDrainTestApp()
	if a.Migrating() {
		t.Fatal("expected app not migrating before Drain")
	}
	called := make(chan struct{})
	a.Drain(false, func() { close(called) })
	if a.Migrating() {
		t.Fatal("expected Migrating false for natural drain")
	}
	waitDrainCallback(t, called)
	a.Drain(true, func() {})
	if a.Migrating() {
		t.Fatal("expected second Drain to be ignored")
	}
}

func TestDrainMigrateWaitsForInstances(t *testing.T) {
	a := newDrainTestApp()
	a.InstanceCreated()
	called := make(chan struct{})
	a.Drain(true, func() { close(called) })
	if !a.Migrating() {
		t.Fatal("expected Migrating after Drain(true)")
	}
	select {
	case <-called:
		t.Fatal("drain callback called before last instance ended")
	case <-time.After(50 * time.Millisecond):
	}
	a.InstanceDeleted()
	waitDrainCallback(t, called)
}
