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
	"testing"
	"time"

	"golang.org/x/text/language"
)

func TestBrowserLocale(t *testing.T) {
	ctx, inst := helperContext(t)
	b := BrowserLocale(ctx)
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	want := Locale{Language: language.MustParse("de-DE"), TimeZone: berlin}
	sessCtx := SessionContext(ctx)
	inst.session.locale.Update(sessCtx, want)
	if got := b.Get(); got != want {
		t.Fatalf("Get: got %+v, want %+v", got, want)
	}
	if got, ok := b.Read(sessCtx); !ok || got != want {
		t.Fatalf("Read: got %+v %v, want %+v", got, ok, want)
	}
}

func TestBrowserLocalePanicsWithoutSession(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	BrowserLocale(context.Background())
}
