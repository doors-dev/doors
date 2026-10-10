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
	"encoding/json"
	"testing"

	"github.com/doors-dev/doors/internal/front/actions"
)

// ActionScroll options that cannot be encoded fail when the action is
// dispatched, not while the update stream writes it.
func TestActionScrollEncodesOptionsUpfront(t *testing.T) {
	if _, err := (ActionScroll{Selector: "#x", Options: func() {}}).action(context.Background(), nil, false); err == nil {
		t.Fatal("expected an encoding error at dispatch")
	}
	prep, err := (ActionScroll{Selector: "#x", Options: map[string]string{"block": "center"}}).action(context.Background(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	options, ok := prep.action.(actions.Scroll).Options.(json.RawMessage)
	if !ok || string(options) != `{"block":"center"}` {
		t.Fatalf("expected encoded options, got %#v", prep.action.(actions.Scroll).Options)
	}
}
