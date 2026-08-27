package app

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestOptionsDefaults(t *testing.T) {
	o := Options{}
	o.initDefaults()
	if o.ID != "doors" {
		t.Fatalf("unexpected default server id: %q", o.ID)
	}
	if o.Conf.ServerSessionCookiePrefix != "" {
		t.Fatalf("unexpected default session cookie prefix: %q", o.Conf.ServerSessionCookiePrefix)
	}
}

type recordingTracker struct {
	name   string
	events *[]string
}

func (t recordingTracker) Create(id string, r *http.Request) {
	*t.events = append(*t.events, t.name+":create:"+id)
}

func (t recordingTracker) Delete(id string) {
	*t.events = append(*t.events, t.name+":delete:"+id)
}

func TestTrackersFanOutInOrder(t *testing.T) {
	var events []string
	all := trackers{
		recordingTracker{name: "first", events: &events},
		recordingTracker{name: "second", events: &events},
	}

	all.Create("s1", httptest.NewRequest(http.MethodGet, "/", nil))
	all.Delete("s1")

	want := []string{
		"first:create:s1",
		"second:create:s1",
		"first:delete:s1",
		"second:delete:s1",
	}
	if !slices.Equal(events, want) {
		t.Fatalf("expected %v, got %v", want, events)
	}
}

func TestTrackersEmptyIsNoOp(t *testing.T) {
	var all trackers
	all.Create("s1", httptest.NewRequest(http.MethodGet, "/", nil))
	all.Delete("s1")
}
