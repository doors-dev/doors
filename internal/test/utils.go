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

package test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	dpath "github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

var Host string

func Float(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

type Bro struct {
	p       int
	b       *rod.Browser
	a       doors.App
	s       *http.Server
	closeCh chan struct{}
	l       net.Listener
}

func LimitMode() bool {
	return os.Getenv("LIMIT") != ""
}

func (s *Bro) App() doors.App {
	return s.a
}

func (s *Bro) Close() {
	s.s.Close()
	<-s.closeCh
	s.l.Close()
}

func (s *Bro) PageStatus(t *testing.T, path string, status int) *rod.Page {
	t.Helper()
	page := s.b.MustPage("")
	var err string
	done := false
	url := s.url(path)
	timed := page.Timeout(pageLoadTimeout)
	defer timed.CancelTimeout()
	wait := timed.EachEvent(
		func(e *proto.NetworkResponseReceived) bool {
			if e.Response.URL != url {
				return false
			}
			if e.Response.Status != status {
				err = fmt.Sprintf("[http %d] %s", int(e.Response.Status), e.Response.URL)
			}
			done = true
			return true
		},
		func(e *proto.NetworkLoadingFailed) bool {
			if e.Canceled {
				return false
			}
			err = fmt.Sprintf("[request-failed] %s – %s", e.RequestID, e.ErrorText)
			done = true
			return true
		},
		func(_ *proto.PageLoadEventFired) bool {
			done = true
			return true
		},
	)
	_ = page.Navigate(url)
	wait()
	if err != "" {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("page load timed out: ", path)
	}
	return page
}
func (s *Bro) Page(t *testing.T, path string) *rod.Page {
	t.Helper()
	page := s.b.MustPage("")
	var err string
	done := false
	timed := page.Timeout(pageLoadTimeout)
	defer timed.CancelTimeout()
	wait := timed.EachEvent(
		func(e *proto.NetworkResponseReceived) bool {
			if e.Response.Status >= 400 {
				err = fmt.Sprintf("[http %d] %s", int(e.Response.Status), e.Response.URL)
				done = true
				return true
			}
			return false
		},
		func(e *proto.NetworkLoadingFailed) bool {
			if e.Canceled {
				return false
			}
			err = fmt.Sprintf("[request-failed] %s – %s", e.RequestID, e.ErrorText)
			done = true
			return true
		},
		func(_ *proto.PageLoadEventFired) bool {
			done = true
			return true
		},
	)
	page.MustNavigate(s.url(path))
	wait()
	if err != "" {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("page load timed out: ", path)
	}
	return page
}

const pageLoadTimeout = 30 * time.Second

func (s *Bro) url(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", s.p, path)
}

func NewFragmentBro(b *rod.Browser, f func() Fragment) *Bro {
	return NewPathBro(b, func(s doors.Source[Path]) gox.Comp {
		return &Page{
			Source: s,
			F:      f(),
		}
	})
}

func NewPathBro(b *rod.Browser, page func(PathLens) gox.Comp, middleware ...doors.Use) *Bro {
	return NewPathBroOptions(
		b,
		page,
		nil,
		middleware...,
	)
}

func NewPathBroOptions(b *rod.Browser, page func(PathLens) gox.Comp, options []doors.With, middleware ...doors.Use) *Bro {
	adapter, err := dpath.GetModelAdapter[Path]()
	if err != nil {
		panic(err)
	}
	return NewBroOptions(
		b,
		func(ctx context.Context, req doors.Request) gox.Comp {
			pathLens := doors.DeriveSourceEqual(doors.Router(ctx), func(l doors.Location) Path {
				m, ok := adapter.Decode(l)
				if !ok {
					return Path{}
				}
				return *m
			}, func(prev doors.Location, m Path) doors.Location {
				loc, err := adapter.Encode(&m)
				if err != nil {
					return prev
				}
				return loc
			}, func(old Path, new Path) bool {
				return reflect.DeepEqual(old, new)
			})
			return page(pathLens)
		},
		options,
		middleware...,
	)
}

func NewBro(browser *rod.Browser, page func(context.Context, doors.Request) gox.Comp, middleware ...doors.Use) *Bro {
	return NewBroOptions(browser, page, nil, middleware...)
}

func NewBroOptions(browser *rod.Browser, page func(context.Context, doors.Request) gox.Comp, options []doors.With, middleware ...doors.Use) *Bro {
	return NewBroWrapOptions(browser, page, nil, options, middleware...)
}

func NewBroWrap(browser *rod.Browser, page func(context.Context, doors.Request) gox.Comp, wrap func(http.Handler) http.Handler, middleware ...doors.Use) *Bro {
	return NewBroWrapOptions(browser, page, wrap, nil, middleware...)
}

func NewBroWrapOptions(browser *rod.Browser, page func(context.Context, doors.Request) gox.Comp, wrap func(http.Handler) http.Handler, options []doors.With, middleware ...doors.Use) *Bro {
	baseOptions := []doors.With{doors.WithID(doors.IDRand())}
	if LimitMode() {
		conf := doors.Conf{}
		conf.InstanceGoroutineLimit = 1
		baseOptions = append(baseOptions, doors.WithConf(conf))
	}
	options = append(baseOptions, options...)
	a := doors.NewApp(page, options...)
	if len(middleware) != 0 {
		a.Use(middleware...)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Error starting listner: %v\n", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	Host = fmt.Sprintf("http://localhost:%d", port)
	println("Started on port", port)
	handler := http.Handler(a)
	if wrap != nil {
		handler = wrap(handler)
	}
	s := &http.Server{
		Handler: handler,
	}
	ch := make(chan struct{}, 0)
	go func() {
		if err := s.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
		close(ch)
	}()
	return &Bro{
		p:       port,
		l:       listener,
		b:       browser,
		a:       a,
		s:       s,
		closeCh: ch,
	}
}

func actionElement(t *testing.T, page *rod.Page, selector string, what string) (*rod.Element, func()) {
	t.Helper()
	p := page.Timeout(200 * time.Millisecond)
	el, err := p.Element(selector)
	p.CancelTimeout()
	if err != nil {
		t.Fatal(what, ": element ", selector, " not found")
	}
	el = el.Context(page.GetContext()).Timeout(waitLimit)
	return el, func() { el.CancelTimeout() }
}

func TestType(t *testing.T, page *rod.Page, selector string, keys []input.Key) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "must")
	defer cancel()
	err := el.Type(keys...)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
	<-time.After(200 * time.Millisecond)
}

func TestInput(t *testing.T, page *rod.Page, selector string, value string) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "must")
	defer cancel()
	err := el.Input(value)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
	<-time.After(200 * time.Millisecond)
}
func TestInputTime(t *testing.T, page *rod.Page, selector string, now time.Time) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "must")
	defer cancel()
	err := el.InputTime(now)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
	<-time.After(200 * time.Millisecond)
}
func TestInputColor(t *testing.T, page *rod.Page, selector string, color string) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "must")
	defer cancel()
	err := el.InputColor(color)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
	<-time.After(200 * time.Millisecond)
}

func TestSelect(t *testing.T, page *rod.Page, selector string, options []string) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "must")
	defer cancel()
	err := el.Select(options, true, rod.SelectorTypeText)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
	<-time.After(200 * time.Millisecond)
}
func TestDeselect(t *testing.T, page *rod.Page, selector string, options []string) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "must")
	defer cancel()
	err := el.Select(options, false, rod.SelectorTypeText)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
	<-time.After(200 * time.Millisecond)
}

const waitLimit = 5 * time.Second

const pollInterval = 20 * time.Millisecond

func poll(page *rod.Page, check func(p *rod.Page) (bool, string)) (bool, string) {
	deadline := time.Now().Add(waitLimit)
	last := ""
	for {
		p := page.Sleeper(rod.NotFoundSleeper).Timeout(time.Until(deadline))
		ok, state := check(p)
		expired := p.GetContext().Err() != nil
		p.CancelTimeout()
		if ok {
			return true, state
		}
		if !expired || last == "" {
			last = state
		}
		if !time.Now().Before(deadline) {
			return false, last
		}
		<-time.After(pollInterval)
	}
}

func TestMust(t *testing.T, page *rod.Page, selector string) {
	t.Helper()
	ok, _ := poll(page, func(p *rod.Page) (bool, string) {
		_, err := p.Element(selector)
		return err == nil, ""
	})
	if !ok {
		t.Fatal("must: element ", selector, " not found")
	}
}

func TestMustNow(t *testing.T, page *rod.Page, selector string) {
	t.Helper()
	p := page.Timeout(200 * time.Millisecond)
	defer p.CancelTimeout()
	_, err := p.Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
}

func TestMustNot(t *testing.T, page *rod.Page, selector string) {
	t.Helper()
	p := page.Timeout(200 * time.Millisecond)
	defer p.CancelTimeout()
	_, err := p.Element(selector)
	if err == nil {
		t.Fatal("must not: element ", selector, " found")
	}
}

func Click(t *testing.T, page *rod.Page, selector string) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "click")
	defer cancel()
	el.MustClick()
	<-time.After(200 * time.Millisecond)
}

func ClickNow(t *testing.T, page *rod.Page, selector string) {
	t.Helper()
	el, cancel := actionElement(t, page, selector, "click")
	defer cancel()
	el.MustClick()
}

func ClickBurst(t *testing.T, page *rod.Page, selectors ...string) {
	t.Helper()
	_, err := page.Eval(`(selectors) => {
		for (const selector of selectors) {
			const el = document.querySelector(selector)
			if (!(el instanceof HTMLElement)) {
				throw new Error("click burst element not found: " + selector)
			}
			el.click()
		}
	}`, selectors)
	if err != nil {
		t.Fatal("click burst failed: ", err)
	}
}

// HijackFailOnce intercepts requests whose URL matches the rod glob pattern and
// fails the first match with a network-level ConnectionFailed error, letting
// every other request (including later matches) pass through untouched. Only the
// given pattern is registered with the CDP Fetch domain, so unrelated traffic
// such as the SSE sync stream, resource loads and page navigations is never
// paused. The returned stop function tears the router down; it is safe to call
// more than once.
func HijackFailOnce(t *testing.T, page *rod.Page, pattern string) func() {
	t.Helper()
	router := page.HijackRequests()
	var failed atomic.Bool
	router.MustAdd(pattern, func(h *rod.Hijack) {
		if failed.CompareAndSwap(false, true) {
			h.Response.Fail(proto.NetworkErrorReasonConnectionFailed)
			return
		}
		h.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()
	var once sync.Once
	return func() {
		once.Do(func() {
			router.MustStop()
		})
	}
}

func TestContent(t *testing.T, page *rod.Page, selector string, content string) {
	t.Helper()
	ok, state := poll(page, func(p *rod.Page) (bool, string) {
		return checkContent(p, selector, content)
	})
	if !ok {
		t.Fatal("content: element ", selector, " ", state)
	}
}

func TestContentNow(t *testing.T, page *rod.Page, selector string, content string) {
	t.Helper()
	p := page.Timeout(300 * time.Millisecond)
	defer p.CancelTimeout()
	if ok, state := checkContent(p, selector, content); !ok {
		t.Fatal("content: element ", selector, " ", state)
	}
}

func checkContent(page *rod.Page, selector string, content string) (bool, string) {
	el, err := page.Element(selector)
	if err != nil {
		return false, "not found"
	}
	s, err := el.Text()
	if err != nil {
		return false, "no text"
	}
	if s != content {
		return false, fmt.Sprint("no exects: ", content, " fact: ", s)
	}
	return true, ""
}

func TestClass(t *testing.T, page *rod.Page, selector string, className string) {
	t.Helper()
	ok, state := poll(page, func(p *rod.Page) (bool, string) {
		return checkClass(p, selector, className)
	})
	if !ok {
		t.Fatal("class: element ", selector, " ", state)
	}
}

func TestClassNow(t *testing.T, page *rod.Page, selector string, className string) {
	t.Helper()
	p := page.Timeout(200 * time.Millisecond)
	defer p.CancelTimeout()
	if ok, state := checkClass(p, selector, className); !ok {
		t.Fatal("class: element ", selector, " ", state)
	}
}

func checkClass(page *rod.Page, selector string, className string) (bool, string) {
	el, err := page.Element(selector)
	if err != nil {
		return false, "not found"
	}
	classAttr, err := el.Attribute("class")
	if err != nil {
		return false, "attribute 'class' not found"
	}
	if classAttr == nil {
		return false, "has no 'class' attribute"
	}
	if !slices.Contains(strings.Fields(*classAttr), className) {
		return false, fmt.Sprint("expects to have class: ", className, " fact: ", *classAttr)
	}
	return true, ""
}

func TestClassNot(t *testing.T, page *rod.Page, selector string, className string) {
	t.Helper()
	page = page.Timeout(200 * time.Millisecond)
	defer page.CancelTimeout()
	el, err := page.Element(selector)
	if err != nil {
		t.Fatal("class: element ", selector, " not found")
	}
	classAttr, err := el.Attribute("class")
	if err != nil {
		t.Fatal("class: element ", selector, " attribute 'class' not found")
	}
	if classAttr == nil {
		return
	}
	classes := strings.Fields(*classAttr)
	found := slices.Contains(classes, className)
	for found {
		t.Fatal("class: element ", selector, " expects not to have class: ", className)
	}
}

func TestAttr(t *testing.T, page *rod.Page, selector string, name string, value string) {
	t.Helper()
	ok, state := poll(page, func(p *rod.Page) (bool, string) {
		return checkAttr(p, selector, name, value)
	})
	if !ok {
		t.Fatal("attr: element ", selector, " ", state)
	}
}

func TestAttrNow(t *testing.T, page *rod.Page, selector string, name string, value string) {
	t.Helper()
	p := page.Timeout(200 * time.Millisecond)
	defer p.CancelTimeout()
	if ok, state := checkAttr(p, selector, name, value); !ok {
		t.Fatal("attr: element ", selector, " ", state)
	}
}

func checkAttr(page *rod.Page, selector string, name string, value string) (bool, string) {
	el, err := page.Element(selector)
	if err != nil {
		return false, "not found"
	}
	attr, err := el.Attribute(name)
	if err != nil {
		return false, fmt.Sprint("attribute ", name, " not found")
	}
	if attr == nil {
		return false, fmt.Sprint("attribute ", name, " is nil")
	}
	if *attr != value {
		return false, fmt.Sprint("attribute ", name, " expects: ", value, " fact: ", *attr)
	}
	return true, ""
}

func TestAttrNo(t *testing.T, page *rod.Page, selector string, name string) {
	t.Helper()
	page = page.Timeout(200 * time.Millisecond)
	defer page.CancelTimeout()
	el, err := page.Element(selector)
	if err != nil {
		t.Fatal("attr: element ", selector, " not found")
	}
	attr, err := el.Attribute(name)
	if err != nil {
		t.Fatal("attr: element ", selector, " attribute ", name, " read error: ", err)
	}

	if attr != nil {
		t.Fatal("attr: element ", selector, " should not have attribute ", name, " fact: ", *attr)
	}
}

func TestAttrGone(t *testing.T, page *rod.Page, selector string, name string) {
	t.Helper()
	ok, state := poll(page, func(p *rod.Page) (bool, string) {
		el, err := p.Element(selector)
		if err != nil {
			return false, fmt.Sprint("element ", selector, " not found")
		}
		attr, err := el.Attribute(name)
		if err != nil {
			return false, fmt.Sprint("attribute ", name, " read error: ", err)
		}
		if attr != nil {
			return false, fmt.Sprint("attribute ", name, " still present: ", *attr)
		}
		return true, ""
	})
	if !ok {
		t.Fatal("attr: element ", selector, " ", state)
	}
}

func TestAttrNot(t *testing.T, page *rod.Page, selector string, name string, value string) {
	t.Helper()
	page = page.Timeout(200 * time.Millisecond)
	defer page.CancelTimeout()
	el, err := page.Element(selector)
	if err != nil {
		t.Fatal("attr: element ", selector, " not found")
	}
	attr, err := el.Attribute(name)
	if err != nil {
		t.Fatal("attr: element ", selector, " attribute ", name, " not found")
	}
	if attr == nil {
		return
	}
	if *attr == value {
		t.Fatal("attr: element ", selector, " attribute ", name, " expects not: ", value)
	}
}

func TestReport(t *testing.T, page *rod.Page, content string) {
	t.Helper()
	TestReportId(t, page, 0, content)
}

func GetContent(t *testing.T, page *rod.Page, selector string) string {
	t.Helper()
	p := page.Timeout(200 * time.Millisecond)
	defer p.CancelTimeout()
	el, err := p.Element(selector)
	if err != nil {
		t.Fatal("content: element ", selector, " not found")
	}
	s, err := el.Text()
	if err != nil {
		t.Fatal("content: element ", selector, " no text")
	}
	return s
}

func GetReportContent(t *testing.T, page *rod.Page, id int) string {
	t.Helper()
	return GetContent(t, page, fmt.Sprintf("#report-%d", id))
}
func TestReportId(t *testing.T, page *rod.Page, id int, content string) {
	t.Helper()
	TestContent(t, page, fmt.Sprintf("#report-%d", id), content)
}

func Text(s string) gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		return cur.Text(s)
	})
}

func Count(page *rod.Page, s string) int {
	elements := page.MustElements(s)
	return len(elements)
}

type RandFile struct {
	Path string
	Hash string // SHA-256 hex digest of the file content
}

func (r *RandFile) IsSame(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		panic(err)
	}
	return r.Hash == hex.EncodeToString(h.Sum(nil))
}

func NewRandFile(size int64) RandFile {
	f, err := os.CreateTemp("", "randfile-*.bin")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h := sha256.New()
	mw := io.MultiWriter(f, h)

	if _, err := io.CopyN(mw, rand.Reader, size); err != nil {
		panic(err)
	}
	if err := f.Sync(); err != nil {
		panic(err)
	}

	return RandFile{
		Path: f.Name(),
		Hash: hex.EncodeToString(h.Sum(nil)),
	}
}
