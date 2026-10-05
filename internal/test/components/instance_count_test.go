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

package components

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

func TestInstanceCount(t *testing.T) {
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			Header: "InstanceCount",
			F:      &InstanceCountFragment{},
		}
	})
	t.Cleanup(bro.Close)

	if got := bro.App().InstanceCount(); got != 0 {
		t.Fatalf("expected 0 instances, got %d", got)
	}

	page1 := bro.Page(t, "/")
	if got := bro.App().InstanceCount(); got != 1 {
		t.Fatalf("expected 1 instance after first page, got %d", got)
	}

	page2 := bro.Page(t, "/")
	if got := bro.App().InstanceCount(); got != 2 {
		t.Fatalf("expected 2 instances after second page, got %d", got)
	}

	test.ClickNow(t, page2, "#end-instance")
	page2.Close()
	<-time.After(500 * time.Millisecond)
	if got := bro.App().InstanceCount(); got != 1 {
		t.Fatalf("expected 1 instance after ending one, got %d", got)
	}

	test.ClickNow(t, page1, "#end-instance")
	page1.Close()
	<-time.After(500 * time.Millisecond)
	if got := bro.App().InstanceCount(); got != 0 {
		t.Fatalf("expected 0 instances after ending one, got %d", got)
	}
}

func TestSessionCount(t *testing.T) {
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			Header: "InstanceCount",
			F:      &InstanceCountFragment{},
		}
	})
	t.Cleanup(bro.Close)

	if got := bro.App().SessionCount(); got != 0 {
		t.Fatalf("expected 0 sessions, got %d", got)
	}

	page := bro.Page(t, "/")
	if got := bro.App().SessionCount(); got != 1 {
		t.Fatalf("expected 1 session after first page, got %d", got)
	}

	test.ClickNow(t, page, "#end-session")
	page.Close()
	<-time.After(500 * time.Millisecond)
	if got := bro.App().SessionCount(); got != 0 {
		t.Fatalf("expected 0 sessions after ending one, got %d", got)
	}
}
