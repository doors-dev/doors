package test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type bro struct {
	r       doors.Router
	s       *http.Server
	closeCh chan struct{}
}

func (s *bro) close() {
	s.s.Close()
	<-s.closeCh
}

func (s *bro) page(t *testing.T, path string) *rod.Page {
	t.Helper()
	page := browser.MustPage("")
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

func (s *bro) url(path string) string {
	return fmt.Sprintf("http://localhost:8088%s", path)
}

func newBro(mods ...doors.Mod) *bro {
	r := doors.NewRouter()
	r.Use(mods...)
	s := &http.Server{
		Addr:    ":8088",
		Handler: r,
	}
	ch := make(chan struct{}, 0)
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
		close(ch)
	}()
	return &bro{
		r:       r,
		s:       s,
		closeCh: ch,
	}
}

func testMust(t *testing.T, page *rod.Page, selector string) {
	_, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
}
func testMustNot(t *testing.T, page *rod.Page, selector string) {
	_, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err == nil {
		t.Fatal("must not: element ", selector, " found")
	}
}

func click(t *testing.T, page *rod.Page, selector string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("click: element ", selector, " not found")
	}
	el.MustClick()
	<-time.After(100 * time.Millisecond)
}

func testContent(t *testing.T, page *rod.Page, selector string, content string) {
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

func testReport(t *testing.T, page *rod.Page, content string) {
	testReportId(t, page, 0, content)
}

func testReportId(t *testing.T, page *rod.Page, id int, content string) {
	testContent(t, page, fmt.Sprintf("#report-%d", id), content)
}
func text(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(s))
		return err
	})
}
