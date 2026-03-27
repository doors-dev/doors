package model

import (
	"errors"
	"testing"
)

func TestInstanceCreationErrorAndResEntity(t *testing.T) {
	errValue := InstanceCreationError{}
	if errValue.Error() != "instance creation error" {
		t.Fatalf("unexpected error string: %q", errValue.Error())
	}

	boom := errors.New("boom")
	res := Res{entity: boom}
	if got := res.Entity(); got != boom {
		t.Fatalf("unexpected entity: %#v", got)
	}
	if !errors.Is(res.Err(), boom) {
		t.Fatalf("unexpected res error: %v", res.Err())
	}
	if _, ok := res.Instance(); ok {
		t.Fatal("expected non-instance result to report no instance")
	}
}
