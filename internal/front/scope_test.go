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

package front

import "testing"

func TestFreeScope(t *testing.T) {
	scope := FreeScope("free-id")
	if scope.Type != "free" {
		t.Fatalf("unexpected scope type: %q", scope.Type)
	}
	if scope.Id != "free-id" {
		t.Fatalf("unexpected scope id: %q", scope.Id)
	}
	if scope.Opt != nil {
		t.Fatalf("expected no scope options, got %#v", scope.Opt)
	}
}
