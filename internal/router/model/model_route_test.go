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
