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
