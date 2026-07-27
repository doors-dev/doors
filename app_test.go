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

package doors

import (
	"net/http"
	"testing"

	"github.com/doors-dev/doors/internal/app"
)

type namedTracker struct {
	name string
}

func (t namedTracker) Create(id string, r *http.Request) {}

func (t namedTracker) Delete(id string) {}

func applyWith(withs ...With) app.Options {
	var o app.Options
	for _, w := range withs {
		w.apply(&o)
	}
	return o
}

func TestWithSessionTrackerAccumulates(t *testing.T) {
	o := applyWith(
		WithSessionTracker(namedTracker{name: "first"}),
		WithSessionTracker(namedTracker{name: "second"}),
	)
	if len(o.SessionTrackers) != 2 {
		t.Fatalf("expected 2 trackers, got %d", len(o.SessionTrackers))
	}
	for i, want := range []string{"first", "second"} {
		got, ok := o.SessionTrackers[i].(namedTracker)
		if !ok {
			t.Fatalf("tracker %d has unexpected type %T", i, o.SessionTrackers[i])
		}
		if got.name != want {
			t.Fatalf("tracker %d: expected %s, got %s", i, want, got.name)
		}
	}
}

func TestWithSessionTrackerSkipsNil(t *testing.T) {
	o := applyWith(
		WithSessionTracker(nil),
		WithSessionTracker(namedTracker{name: "only"}),
		WithSessionTracker(nil),
	)
	if len(o.SessionTrackers) != 1 {
		t.Fatalf("expected nil trackers to be skipped, got %d", len(o.SessionTrackers))
	}
}
