package app

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
)

func newCookieTestApp(prefix string, serverIDCookie string) *app {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	conf.ServerSessionCookiePrefix = prefix
	return &app{
		conf:      conf,
		pathMaker: path.NewPathMaker(prefix, "test", serverIDCookie),
	}
}

func TestSetCookiesSessionCookie(t *testing.T) {
	a := newCookieTestApp("__Host-", "")
	w := httptest.NewRecorder()
	a.SetCookies(w, "sid", time.Minute)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one session cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "__Host-test" {
		t.Fatalf("unexpected cookie name: %q", cookie.Name)
	}
	if cookie.Value != "sid" {
		t.Fatalf("unexpected cookie value: %q", cookie.Value)
	}
	if !cookie.Secure {
		t.Fatal("expected session cookie to be secure")
	}
	if !cookie.HttpOnly {
		t.Fatal("expected session cookie to be HttpOnly")
	}
	if cookie.Path != "/" {
		t.Fatalf("unexpected cookie path: %q", cookie.Path)
	}
	if cookie.MaxAge != 60 {
		t.Fatalf("unexpected cookie max age: %d", cookie.MaxAge)
	}

	a = newCookieTestApp("", "")
	a.conf.ServerSessionCookieNoSecure = true
	w = httptest.NewRecorder()
	a.SetCookies(w, "sid", time.Minute)
	cookies = w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one no-secure session cookie, got %d", len(cookies))
	}
	cookie = cookies[0]
	if cookie.Name != "test" {
		t.Fatalf("unexpected no-secure cookie name: %q", cookie.Name)
	}
	if cookie.Secure {
		t.Fatal("expected no-secure session cookie to omit Secure")
	}
}

func TestSetCookiesServerIDCookie(t *testing.T) {
	a := newCookieTestApp("", "server_id")
	w := httptest.NewRecorder()
	a.SetCookies(w, "sid", time.Minute)
	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected two cookies (session + server ID), got %d", len(cookies))
	}
	serverCookie := cookies[1]
	if serverCookie.Name != "server_id" {
		t.Fatalf("unexpected server ID cookie name: %q", serverCookie.Name)
	}
	if serverCookie.Value != "test" {
		t.Fatalf("unexpected server ID cookie value: %q", serverCookie.Value)
	}
	if !serverCookie.HttpOnly {
		t.Fatal("expected server ID cookie to be HttpOnly")
	}
	if !serverCookie.Secure {
		t.Fatal("expected server ID cookie to be Secure")
	}
	if serverCookie.Path != "/" {
		t.Fatalf("unexpected server ID cookie path: %q", serverCookie.Path)
	}
	if serverCookie.MaxAge != 60 {
		t.Fatalf("unexpected server ID cookie max age: %d", serverCookie.MaxAge)
	}
}
