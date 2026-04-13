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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResourceHook(t *testing.T) {
	called := false
	resource := ResourceHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		called = true
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
		return true
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)

	done := resource.Handler()(context.Background(), recorder, request)
	if !done {
		t.Fatal("expected resource hook to request shutdown")
	}
	if !called {
		t.Fatal("expected resource hook handler to be called")
	}
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if recorder.Body.String() != "ok" {
		t.Fatalf("unexpected body: %q", recorder.Body.String())
	}
}
