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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/common"
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
	r       doors.Router
	s       *http.Server
	closeCh chan struct{}
	l       net.Listener
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
	url := s.url(path)
	wait := page.EachEvent(
		func(e *proto.NetworkResponseReceived) bool {
			if e.Response.URL != url {
				return false
			}
			if e.Response.Status != status {
				err = fmt.Sprintf("[http %d] %s", int(e.Response.Status), e.Response.URL)
			}
			return true
		},
		func(e *proto.NetworkLoadingFailed) bool {
			err = fmt.Sprintf("[request-failed] %s – %s", e.RequestID, e.ErrorText)
			return true
		},
		func(_ *proto.PageLoadEventFired) bool {
			return true
		},
	)
	page.MustNavigate(url)
	wait()
	if err != "" {
		t.Fatal(err)
	}
	return page
}
func (s *Bro) Page(t *testing.T, path string) *rod.Page {
	t.Helper()
	page := s.b.MustPage("")
	var err string
	wait := page.EachEvent(
		func(e *proto.NetworkResponseReceived) bool {
			if e.Response.Status >= 400 {
				err = fmt.Sprintf("[http %d] %s", int(e.Response.Status), e.Response.URL)
				return true
			}
			return false
		},
		func(e *proto.NetworkLoadingFailed) bool {
			err = fmt.Sprintf("[request-failed] %s – %s", e.RequestID, e.ErrorText)
			return true
		},
		func(_ *proto.PageLoadEventFired) bool {
			return true
		},
	)
	page.MustNavigate(s.url(path))
	wait()
	if err != "" {
		t.Fatal(err)
	}
	return page
}

func (s *Bro) url(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", s.p, path)
}

func NewFragmentBro(b *rod.Browser, f func() Fragment) *Bro {
	return NewBro(b,
		doors.ServePage(func(pr doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
			return pr.Page(&Page{
				F: f(),
			})
		}),
	)
}


func NewBro(browser *rod.Browser, mods ...doors.Mod) *Bro {
	r := doors.NewRouter()
	r.Use(mods...)
	limit := os.Getenv("LIMIT") != ""
	if limit {
		r.Use(doors.SetSystemConf(common.SystemConf{
			SessionInstanceLimit:   1,
			InstanceGoroutineLimit: 1,
		}))
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Error starting listner: %v\n", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	Host = fmt.Sprintf("http://localhost:%d", port)
	println("Started on port", port)
	s := &http.Server{
		Handler: r,
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
		r:       r,
		s:       s,
		closeCh: ch,
	}
}
func TestType(t *testing.T, page *rod.Page, selector string, keys []input.Key) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
	err = el.Type(keys...)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
}

func TestInput(t *testing.T, page *rod.Page, selector string, value string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
	err = el.Input(value)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
}
func TestInputTime(t *testing.T, page *rod.Page, selector string, now time.Time) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
	err = el.InputTime(now)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
}
func TestInputColor(t *testing.T, page *rod.Page, selector string, color string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
	err = el.InputColor(color)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
}

func TestSelect(t *testing.T, page *rod.Page, selector string, options []string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
	err = el.Select(options, true, rod.SelectorTypeText)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
}
func TestDeselect(t *testing.T, page *rod.Page, selector string, options []string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
	err = el.Select(options, false, rod.SelectorTypeText)
	if err != nil {
		t.Fatal("must: element ", selector, " input failed")
	}
}

func TestMust(t *testing.T, page *rod.Page, selector string) {
	_, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
}
func TestMustNot(t *testing.T, page *rod.Page, selector string) {
	_, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err == nil {
		t.Fatal("must not: element ", selector, " found")
	}
}

func Click(t *testing.T, page *rod.Page, selector string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("click: element ", selector, " not found")
	}
	el.MustClick()
	<-time.After(100 * time.Millisecond)
}

func ClickNow(t *testing.T, page *rod.Page, selector string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("click: element ", selector, " not found")
	}
	el.MustClick()
}

func TestContent(t *testing.T, page *rod.Page, selector string, content string) {
	page = page.Timeout(200 * time.Millisecond)
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("content: element ", selector, " not found")
	}
	s, err := el.Text()
	if err != nil {
		t.Fatal("content: element ", selector, " no text")
	}
	if s != content {
		t.Fatal("content: element ", selector, " no exects: ", content, " fact: ", s)
	}
}

func TestClass(t *testing.T, page *rod.Page, selector string, className string) {
	page = page.Timeout(200 * time.Millisecond)
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("class: element ", selector, " not found")
	}
	classAttr, err := el.Attribute("class")
	if err != nil {
		t.Fatal("class: element ", selector, " attribute 'class' not found")
	}
	if classAttr == nil {
		t.Fatal("class: element ", selector, " has no 'class' attribute")
	}
	classes := strings.Fields(*classAttr)
	found := slices.Contains(classes, className)
	if !found {
		t.Fatal("class: element ", selector, " expects to have class: ", className, " fact: ", *classAttr)
	}
}

func TestClassNot(t *testing.T, page *rod.Page, selector string, className string) {
	page = page.Timeout(200 * time.Millisecond)
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
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
	page = page.Timeout(200 * time.Millisecond)
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("attr: element ", selector, " not found")
	}
	attr, err := el.Attribute(name)
	if err != nil {
		t.Fatal("attr: element ", selector, " attribute ", name, " not found")
	}
	if attr == nil {
		t.Fatal("attr: element ", selector, " attribute ", name, " is nil")
	}
	if *attr != value {
		t.Fatal("attr: element ", selector, " attribute ", name, " expects: ", value, " fact: ", *attr)
	}
}

func TestAttrNo(t *testing.T, page *rod.Page, selector string, name string) {
	page = page.Timeout(200 * time.Millisecond)
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
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

func TestAttrNot(t *testing.T, page *rod.Page, selector string, name string, value string) {
	page = page.Timeout(200 * time.Millisecond)
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
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
	TestReportId(t, page, 0, content)
}

func GetContent(t *testing.T, page *rod.Page, selector string) string {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
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
	page = page.Timeout(200 * time.Millisecond)
	selector := fmt.Sprintf("#report-%d", id)
	return GetContent(t, page, selector)
}
func TestReportId(t *testing.T, page *rod.Page, id int, content string) {
	TestContent(t, page, fmt.Sprintf("#report-%d", id), content)
}
func Text(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(s))
		return err
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
