// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/doors-dev/doors/internal/ctex"
)

// ReqAfter allows setting client-side actions to run after a request completes.
type ReqAfter interface {
	// After sets client-side actions to run once the request finishes.
	After([]Action) error
}

// Req provides basic request operations including cookie management.
type Req interface {
	// SetCookie adds a cookie to the response.
	SetCookie(cookie *http.Cookie)
	// GetCookie retrieves a cookie by name.
	GetCookie(name string) (*http.Cookie, error)
	// Done signals when the request context is canceled or completed.
	Done() <-chan struct{}
}

// ReqEvent provides request handling for event hooks with typed event data.
type ReqEvent[E any] interface {
	Req
	ReqAfter
	// Event returns the event payload.
	Event() E
}

// RForm provides request handling for form submissions with typed form data.
type RForm[D any] interface {
	Req
	ReqAfter
	// Data returns the parsed form payload.
	Data() D
}

// ReqRawForm provides access to raw multipart form data for streaming or custom parsing.
type ReqRawForm interface {
	Req
	ReqAfter
	// W returns the HTTP response writer.
	W() http.ResponseWriter
	// Reader returns a multipart reader for streaming form parts.
	Reader() (*multipart.Reader, error)
	// ParseForm parses the form data with a memory limit.
	ParseForm(maxMemory int) (ParsedForm, error)
}

// ParsedForm exposes parsed form values and uploaded files.
type ParsedForm interface {
	// FormValues returns all parsed form values.
	FormValues() url.Values
	// FormValue returns the first value for the given key.
	FormValue(key string) string
	// FormFile returns the uploaded file for the given key.
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	// Form returns the underlying multipart.Form.
	Form() *multipart.Form
}

// ReqHook provides request handling for hook handlers with typed data.
type ReqHook[D any] interface {
	Req
	ReqAfter
	// Data returns the parsed hook payload.
	Data() D
}

// ReqRawHook provides access to raw request data for hook handlers without parsing.
type ReqRawHook interface {
	ReqRawForm
	// Body returns the raw request body reader.
	Body() io.ReadCloser
}

type ReqModel interface {
	Req
	// SessionStore returns session-scoped storage.
	SessionStore() Store
	// RequestHeader returns the incoming request headers.
	RequestHeader() http.Header
	// ResponseHeader returns the outgoing response headers.
	ResponseHeader() http.Header
}

type req struct {
	w   http.ResponseWriter
	r   *http.Request
	ctx context.Context
}

func (r *req) After(action []Action) error {
	actions := intoActions(r.ctx, action)
	err := actions.Set(r.w.Header())
	if err != nil {
		panic(err)
	}
	return nil
}

func (r *req) Body() io.ReadCloser {
	return r.r.Body
}

func (r *req) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.w, cookie)
}

func (r *req) GetCookie(name string) (*http.Cookie, error) {
	return r.r.Cookie(name)
}

func (r *req) ParseForm(maxMemory int) (ParsedForm, error) {
	if maxMemory <= 0 {
		maxMemory = defaultMaxMemory
	}
	return r, r.r.ParseMultipartForm(int64(maxMemory))
}

func (r *req) Reader() (*multipart.Reader, error) {
	return r.r.MultipartReader()
}

func (r *req) FormValues() url.Values {
	return r.r.Form
}

func (r *req) Done() <-chan struct{} {
	return r.r.Context().Done()
}

func (r *req) Form() *multipart.Form {
	return r.r.MultipartForm
}

func (r *req) FormValue(key string) string {
	return r.r.FormValue(key)
}

func (r *req) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return r.r.FormFile(key)
}

func (r *req) W() http.ResponseWriter {
	return r.w
}

type eventRequest[E any] struct {
	req
	e *E
}

func (e *eventRequest[E]) Event() E {
	return *e.e
}

type formHookRequest[D any] struct {
	req
	data *D
}

func (d *formHookRequest[D]) Data() D {
	return *d.data
}

type modelRequest struct {
	req
	store ctex.Store
}

func (r *modelRequest) RequestHeader() http.Header {
	return r.req.r.Header
}

func (r *modelRequest) ResponseHeader() http.Header {
	return r.req.w.Header()
}

func (r *modelRequest) SessionStore() Store {
	return r.store
}
