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
	"io"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
)

type rawAttrValue string

func (r rawAttrValue) Output(w io.Writer) error {
	_, err := io.WriteString(w, string(r))
	return err
}

func TestSetterMutateValueFails(t *testing.T) {
	ctx, _ := helperContext(t)
	c := ctx.Value(common.KeyCore).(core.Core)
	s := &Setter{}
	if _, err := s.Set("class", Class("hl")).action(ctx, c, false); err == nil {
		t.Fatal("expected gox.Mutate value to fail the action")
	}
	if _, err := s.Set("class", "hl").action(ctx, c, false); err != nil {
		t.Fatal(err)
	}
}
