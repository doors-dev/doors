package test

import (
	"testing"

	"github.com/doors-dev/doors"
)

func TestTest(t *testing.T) {
	r := doors.NewRouter()

	if r == nil {
		t.Error("Router is nill")
	}
}
