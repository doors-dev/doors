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
