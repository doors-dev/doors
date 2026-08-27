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
	"fmt"
	"strings"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

func attrValue(t *testing.T, attrs gox.Attrs, name string) string {
	t.Helper()
	attr, ok := attrs.Find(name)
	if !ok {
		t.Fatalf("attribute %s not set", name)
	}
	var b strings.Builder
	if err := attr.OutputValue(&b); err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(b.String())
}

func TestEmitterMultiAttach(t *testing.T) {
	ctx, _ := helperContext(t)
	c := ctx.Value(common.KeyCore).(core.Core)
	a := &Emitter{}
	b := &Emitter{}
	attrs := gox.NewAttrs()
	front.AttrsSetParent(attrs, 7)
	attrs.AddMod(a)
	attrs.AddMod(A(ctx, b))
	if err := attrs.ApplyMods(ctx, "div"); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("[%d,%d]", a.id.ID(c), b.id.ID(c))
	if got := attrValue(t, attrs, "data-d0e"); got != want {
		t.Fatalf("expected emitter attribute %s, got %s", want, got)
	}
	if got := attrValue(t, attrs, "data-d0p"); got != "7" {
		t.Fatalf("expected parent attribute 7, got %s", got)
	}
}

func TestSetterMultiAttach(t *testing.T) {
	ctx, _ := helperContext(t)
	c := ctx.Value(common.KeyCore).(core.Core)
	a := &Setter{}
	b := &Setter{}
	attrs := gox.NewAttrs()
	front.AttrsSetParent(attrs, 7)
	attrs.AddMod(a)
	attrs.AddMod(A(ctx, b))
	if err := attrs.ApplyMods(ctx, "div"); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("[%d,%d]", a.id.ID(c), b.id.ID(c))
	if got := attrValue(t, attrs, "data-d0s"); got != want {
		t.Fatalf("expected setter attribute %s, got %s", want, got)
	}
	if got := attrValue(t, attrs, "data-d0p"); got != "7" {
		t.Fatalf("expected parent attribute 7, got %s", got)
	}
}
