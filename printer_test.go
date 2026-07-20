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
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/doors-dev/gox"
)

type pageStamp struct {
	next  gox.Printer
	count int
}

func (s *pageStamp) Send(j gox.Job) error {
	if open, ok := j.(*gox.JobHeadOpen); ok && open.Kind != gox.KindContainer && open.Attrs != nil {
		s.count++
		open.Attrs.Get("data-render-index").Set(strconv.Itoa(s.count))
	}
	return s.next.Send(j)
}

func TestWithPrinterMiddlewarePageRender(t *testing.T) {
	invoked := 0
	app := NewApp(func(context.Context, Request) gox.Comp {
		return testPage("stamped-page")
	}, WithPrinterMiddleware(func(next gox.Printer) gox.Printer {
		invoked++
		return &pageStamp{next: next}
	}))

	server := httptest.NewServer(app)
	defer server.Close()

	status, _, body := readURL(t, server, "/")
	if status != http.StatusOK {
		t.Fatalf("unexpected page status: %d", status)
	}
	if !strings.Contains(body, "stamped-page") {
		t.Fatalf("expected page content in response, got %q", body)
	}
	if !strings.Contains(body, `data-render-index="1"`) || !strings.Contains(body, `data-render-index="2"`) {
		t.Fatalf("expected middleware-injected attributes in response, got %q", body)
	}
	if invoked != 1 {
		t.Fatalf("expected middleware to be invoked once per drain, got %d", invoked)
	}
}

func TestWithoutPrinterMiddlewarePageRender(t *testing.T) {
	app := NewApp(func(context.Context, Request) gox.Comp {
		return testPage("plain-page")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	status, _, body := readURL(t, server, "/")
	if status != http.StatusOK || !strings.Contains(body, "plain-page") {
		t.Fatalf("expected plain page render, got status=%d body=%q", status, body)
	}
}
