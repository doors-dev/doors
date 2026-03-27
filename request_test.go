package doors

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/doors-dev/doors/internal/ctex"
)

func newMultipartRequest(t *testing.T) (*http.Request, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", "alex"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("file", "hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("payload")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, writer.Boundary()
}

func TestRequestMultipartAndCookies(t *testing.T) {
	req, _ := newMultipartRequest(t)
	req.AddCookie(&http.Cookie{Name: "session", Value: "cookie-1"})
	rec := httptest.NewRecorder()
	r := &request{w: rec, r: req}

	cookie, err := r.GetCookie("session")
	if err != nil {
		t.Fatal(err)
	}
	if cookie.Value != "cookie-1" {
		t.Fatalf("unexpected cookie value: %q", cookie.Value)
	}

	r.SetCookie(&http.Cookie{Name: "written", Value: "cookie-2"})
	if rec.Header().Get("Set-Cookie") == "" {
		t.Fatal("expected Set-Cookie header to be written")
	}

	parsed, err := r.ParseForm(0)
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.FormValue("name"); got != "alex" {
		t.Fatalf("unexpected parsed form value: %q", got)
	}
	if got := r.FormValue("name"); got != "alex" {
		t.Fatalf("unexpected request form value: %q", got)
	}
	if got := parsed.FormValues().Get("name"); got != "alex" {
		t.Fatalf("unexpected parsed values map entry: %q", got)
	}
	if got := r.FormValues().Get("name"); got != "alex" {
		t.Fatalf("unexpected request values map entry: %q", got)
	}
	if parsed.Form() == nil || r.Form() == nil {
		t.Fatal("expected multipart form to be available")
	}

	file, header, err := parsed.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if header.Filename != "hello.txt" {
		t.Fatalf("unexpected uploaded file name: %q", header.Filename)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "payload" {
		t.Fatalf("unexpected uploaded file content: %q", string(content))
	}

	if r.W() != rec {
		t.Fatal("expected request writer to match recorder")
	}
}

func TestRequestReaderBodyDoneAndWrappers(t *testing.T) {
	req, _ := newMultipartRequest(t)
	rec := httptest.NewRecorder()
	r := &request{w: rec, r: req}

	reader, err := r.Reader()
	if err != nil {
		t.Fatal(err)
	}
	part, err := reader.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if part.FormName() != "name" {
		t.Fatalf("unexpected first multipart field: %q", part.FormName())
	}
	_ = part.Close()

	rawReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("raw"))
	rawRec := httptest.NewRecorder()
	rawRequest := &request{w: rawRec, r: rawReq}
	body, err := io.ReadAll(rawRequest.Body())
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "raw" {
		t.Fatalf("unexpected raw body: %q", string(body))
	}

	ctx, cancel := context.WithCancel(context.Background())
	doneReq := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	doneRequest := &request{w: httptest.NewRecorder(), r: doneReq}
	cancel()
	select {
	case <-doneRequest.Done():
	default:
		t.Fatal("expected done channel to close after cancel")
	}

	eventReq := &eventRequest[string]{e: ptr("click")}
	if eventReq.Event() != "click" {
		t.Fatal("expected event wrapper to expose payload")
	}

	formReq := &formHookRequest[string]{data: ptr("value")}
	if formReq.Data() != "value" {
		t.Fatal("expected form hook wrapper to expose data")
	}

	store := ctex.NewStore()
	modelReq := &modelRequest{
		request: request{w: rec, r: httptest.NewRequest(http.MethodGet, "/", nil)},
		store:   store,
	}
	modelReq.RequestHeader().Set("X-Test", "1")
	if modelReq.RequestHeader().Get("X-Test") != "1" {
		t.Fatal("expected request header to be writable")
	}
	modelReq.ResponseHeader().Set("X-Response", "2")
	if rec.Header().Get("X-Response") != "2" {
		t.Fatal("expected response header to be writable")
	}
	if modelReq.SessionStore() != store {
		t.Fatal("expected session store to match model store")
	}
}

func ptr[T any](v T) *T {
	return &v
}
