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

package locale

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func localeBro() *test.Bro {
	return test.NewFragmentBro(browser, func() test.Fragment {
		return &localeFragment{}
	})
}

// localePage opens the app in a page whose navigator.language (fr-FR) and
// time zone differ from the Accept-Language header of its document requests
// (de-DE). The user agent override sets both navigator.language and the
// header. Extra HTTP headers do not reach navigation requests, so the document
// request header is rewritten through request interception.
func localePage(t *testing.T, timeZone string) (*rod.Page, func()) {
	t.Helper()
	page := browser.MustPage("")
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent:      browser.MustVersion().UserAgent,
		AcceptLanguage: "fr-FR",
	})
	if err := (proto.EmulationSetTimezoneOverride{TimezoneID: timeZone}).Call(page); err != nil {
		t.Fatal(err)
	}
	router := page.HijackRequests()
	err := router.Add("*", proto.NetworkResourceTypeDocument, func(h *rod.Hijack) {
		headers := []*proto.FetchHeaderEntry{{Name: "Accept-Language", Value: "de-DE"}}
		for name, value := range h.Request.Headers() {
			if !strings.EqualFold(name, "Accept-Language") {
				headers = append(headers, &proto.FetchHeaderEntry{Name: name, Value: value.Str()})
			}
		}
		h.ContinueRequest(&proto.FetchContinueRequest{Headers: headers})
	})
	if err != nil {
		t.Fatal(err)
	}
	go router.Run()
	page.MustNavigate(test.Host + "/").MustWaitLoad()
	return page, func() {
		router.MustStop()
		page.MustClose()
	}
}

func reload(t *testing.T, page *rod.Page) {
	t.Helper()
	before := test.GetContent(t, page, "#instance-id")
	if err := page.Reload(); err != nil {
		t.Fatal(err)
	}
	page.MustWaitLoad()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if test.GetContent(t, page, "#instance-id") != before {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected a new instance after reload")
}

func htmlContent(t *testing.T, html string, id string) string {
	t.Helper()
	_, after, ok := strings.Cut(html, `id="`+id+`"`)
	if !ok {
		t.Fatalf("element %q not found in initial html", id)
	}
	_, after, _ = strings.Cut(after, ">")
	value, _, _ := strings.Cut(after, "<")
	return value
}

func TestBrowserLocaleAcceptLanguage(t *testing.T) {
	bro := localeBro()
	defer bro.Close()
	req, err := http.NewRequest(http.MethodGet, test.Host+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept-Language", "de;q=0.8, *, fr-CH, it, en;q=0.9")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	if got := htmlContent(t, html, "initial"); got != "fr-CH|local|false" {
		t.Fatalf("expected initial fr-CH|local|false, got %q", got)
	}
	if got := htmlContent(t, html, "live"); got != "fr-CH|local|false" {
		t.Fatalf("expected live fr-CH|local|false, got %q", got)
	}
}

func TestBrowserLocaleReport(t *testing.T) {
	bro := localeBro()
	defer bro.Close()
	page, closePage := localePage(t, "Asia/Tokyo")
	defer closePage()
	test.TestContent(t, page, "#live", "fr-FR|Asia/Tokyo|true")
	test.TestContentNow(t, page, "#initial", "de-DE|local|false")
}

func TestBrowserLocaleReload(t *testing.T) {
	bro := localeBro()
	defer bro.Close()
	page, closePage := localePage(t, "Asia/Tokyo")
	defer closePage()
	test.TestContent(t, page, "#live", "fr-FR|Asia/Tokyo|true")
	reload(t, page)
	test.TestContentNow(t, page, "#initial", "fr-FR|Asia/Tokyo|true")
	test.TestContentNow(t, page, "#live", "fr-FR|Asia/Tokyo|true")
}

func TestBrowserLocaleOtherPageReport(t *testing.T) {
	bro := localeBro()
	defer bro.Close()
	first, closeFirst := localePage(t, "Asia/Tokyo")
	defer closeFirst()
	test.TestContent(t, first, "#live", "fr-FR|Asia/Tokyo|true")
	second, closeSecond := localePage(t, "Europe/Berlin")
	defer closeSecond()
	test.TestContentNow(t, second, "#initial", "fr-FR|Asia/Tokyo|true")
	test.TestContent(t, second, "#live", "fr-FR|Europe/Berlin|true")
	test.TestContent(t, first, "#live", "fr-FR|Europe/Berlin|true")
}
