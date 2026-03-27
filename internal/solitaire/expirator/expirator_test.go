package expirator

import (
	"testing"
	"time"
)

type stubExpirationHandler struct {
	expired int
}

func (s *stubExpirationHandler) Expire() {
	s.expired += 1
}

func TestExpiratorTrackReportAndShutdown(t *testing.T) {
	handler := &stubExpirationHandler{}
	exp := NewExpirator(handler)

	exp.Track(1, time.Now().Add(time.Hour))
	if exp.head == nil || exp.head.id != 1 {
		t.Fatalf("unexpected head after first track: %#v", exp.head)
	}

	exp.Track(2, time.Now().Add(2*time.Hour))
	if exp.lookup[2] == nil {
		t.Fatal("expected later expiration to stay in lookup")
	}

	exp.Report(1)
	if exp.head == nil || exp.head.id != 2 {
		t.Fatalf("unexpected head after reporting first expiration: %#v", exp.head)
	}

	exp.Report(2)
	if exp.head != nil {
		t.Fatalf("expected head to be cleared, got %#v", exp.head)
	}

	exp.Track(3, time.Now().Add(time.Hour))
	exp.Shutdown()
	exp.Shutdown()
	if !exp.expired.Load() {
		t.Fatal("expected shutdown to mark expirator as expired")
	}
}

func TestExpiratorExpireAndHeadReset(t *testing.T) {
	handler := &stubExpirationHandler{}
	exp := NewExpirator(handler)

	exp.newHead(&expiration{id: 7, time: time.Now().Add(time.Hour)})
	if exp.timer == nil {
		t.Fatal("expected timer to be created for head expiration")
	}

	exp.newHead(nil)
	if exp.timer != nil {
		t.Fatal("expected timer to be cleared with nil head")
	}

	exp.expire()
	exp.expire()
	if handler.expired != 1 {
		t.Fatalf("expected expire handler to fire once, got %d", handler.expired)
	}
}
