package test

import (
	"fmt"
	"log"
	"net/http"
	"testing"

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
