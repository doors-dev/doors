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

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func lifecyclePage(t *testing.T) *rod.Page {
	t.Helper()
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
			return doors.ResponseComp(&test.Page{
				Source: r,
				Header: "Lifecycle",
				F:      &LifecycleFragment{},
			})
		})
	})
	t.Cleanup(func() {
		bro.Close()
	})

	page := bro.Page(t, "/")
	t.Cleanup(func() {
		page.Close()
	})
	return page
}

func waitForContentChange(t *testing.T, page *rod.Page, selector string, previous string) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		el, err := page.Timeout(200 * time.Millisecond).Element(selector)
		if err == nil {
			text, err := el.Text()
			if err == nil && text != previous {
				return text
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("expected %s content to change from %q", selector, previous)
	return ""
}

func TestInstanceLifecycle(t *testing.T) {
	page := lifecyclePage(t)

	sessionID := test.GetContent(t, page, "#session-id")
	instanceID := test.GetContent(t, page, "#instance-id")
	sessionMarker := test.GetContent(t, page, "#session-marker")
	instanceMarker := test.GetContent(t, page, "#instance-marker")

	test.ClickNow(t, page, "#end-instance")
	newInstanceMarker := waitForContentChange(t, page, "#instance-marker", instanceMarker)

	newSessionID := test.GetContent(t, page, "#session-id")
	newInstanceID := test.GetContent(t, page, "#instance-id")
	newSessionMarker := test.GetContent(t, page, "#session-marker")

	if newSessionID != sessionID {
		t.Fatal("expected session id to stay stable after instance end")
	}
	if newInstanceID == instanceID {
		t.Fatal("expected instance id to change after instance end")
	}
	if newSessionMarker != sessionMarker {
		t.Fatal("expected session marker to stay stable after instance end")
	}
	if newInstanceMarker == instanceMarker {
		t.Fatal("expected instance marker to change after instance end")
	}
}

func TestSessionLifecycle(t *testing.T) {
	page := lifecyclePage(t)

	sessionID := test.GetContent(t, page, "#session-id")
	instanceID := test.GetContent(t, page, "#instance-id")
	sessionMarker := test.GetContent(t, page, "#session-marker")
	instanceMarker := test.GetContent(t, page, "#instance-marker")

	test.ClickNow(t, page, "#end-session")
	newSessionMarker := waitForContentChange(t, page, "#session-marker", sessionMarker)

	newSessionID := test.GetContent(t, page, "#session-id")
	newInstanceID := test.GetContent(t, page, "#instance-id")
	newInstanceMarker := test.GetContent(t, page, "#instance-marker")

	if newSessionID == sessionID {
		t.Fatal("expected session id to change after session end")
	}
	if newInstanceID == instanceID {
		t.Fatal("expected instance id to change after session end")
	}
	if newSessionMarker == sessionMarker {
		t.Fatal("expected session marker to change after session end")
	}
	if newInstanceMarker == instanceMarker {
		t.Fatal("expected instance marker to change after session end")
	}
}
