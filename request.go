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

package doors

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// RequestAfter schedules client-side actions to run after a successful
// request.
type RequestAfter interface {
	// After appends client-side actions to the successful response.
	After(a Actions) error
}

// RequestCommon exposes the common server-side request helpers available to Doors
// handlers.
type RequestCommon interface {
	// SetCookie adds a cookie to the response.
	SetCookie(cookie *http.Cookie)
	// GetCookie retrieves a cookie by name.
	GetCookie(name string) (*http.Cookie, error)
	// Context returns the underlying HTTP request context.
	// It is not the Doors runtime context passed to handlers.
	Context() context.Context
	// RequestHeader returns the incoming request headers.
	RequestHeader() http.Header
	// RemoteAddr returns the network address that sent the request
	RemoteAddr() string
}

// RequestEvent is the request context passed to event handlers.
type RequestEvent[E any] interface {
	RequestCommon
	RequestAfter
	// Event returns the event payload.
	Event() E
}

// RequestForm is the request context passed to decoded form handlers.
type RequestForm[D any] interface {
	RequestCommon
	RequestAfter
	// Data returns the parsed form payload.
	Data() D
}

// RequestRawForm is the request context passed to raw multipart form handlers.
type RequestRawForm interface {
	RequestCommon
	RequestAfter
	// SetRequestBodyLimit sets the max request body size in bytes for
	// ParseForm, Reader, and Body. Negative values disable the limit, zero
	// forbids body reads. Defaults to the configured server request body
	// limit.
	SetRequestBodyLimit(limit int)
	// ResponseWriter returns the HTTP response writer.
	ResponseWriter() http.ResponseWriter
	// Reader returns a multipart reader for streaming form parts. The body
	// is bounded by the most recently set request body limit.
	Reader() (*multipart.Reader, error)
	// ParseForm parses the form data with a memory limit.
	// The body is bounded by the request body limit.
	ParseForm(maxMemory int) (ParsedForm, error)
}

// ParsedForm exposes parsed multipart form values and files.
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

// RequestHook is the request context passed to typed JavaScript hook handlers.
type RequestHook[D any] interface {
	RequestCommon
	RequestAfter
	// Data returns the parsed hook payload.
	Data() D
}

// RequestRawHook is the request context passed to raw JavaScript hook handlers.
type RequestRawHook interface {
	RequestRawForm
	// Body returns the raw request body reader.
	Body() io.ReadCloser
}

// Request is the request context passed to main app handler.
//
// Use it for cookies, request headers, and response headers.
type Request interface {
	RequestCommon
	// ResponseHeader returns the outgoing response headers.
	ResponseHeader() http.Header
}

type request struct {
	w            http.ResponseWriter
	r            *http.Request
	ctx          context.Context
	limit        int
	defaultLimit int
}

func (r *request) RemoteAddr() string {
	return r.r.RemoteAddr
}

func (r *request) Context() context.Context {
	return r.r.Context()
}

func (r *request) After(a Actions) error {
	actions := intoActions(r.ctx, a.Actions())
	err := actions.Set(r.w.Header())
	if err != nil {
		panic(err)
	}
	return nil
}

func (r *request) Body() io.ReadCloser {
	if r.limit < 0 {
		return r.r.Body
	}
	return http.MaxBytesReader(r.w, r.r.Body, int64(r.limit))
}

func (r *request) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.w, cookie)
}

func (r *request) GetCookie(name string) (*http.Cookie, error) {
	return r.r.Cookie(name)
}

func (r *request) SetRequestBodyLimit(limit int) {
	r.limit = limit
}

func (r *request) ParseForm(maxMemory int) (ParsedForm, error) {
	if r.limit < 0 {
		return r, r.r.ParseMultipartForm(int64(maxMemory))
	}
	r.r.Body = http.MaxBytesReader(r.w, r.r.Body, int64(r.limit))
	return r, r.r.ParseMultipartForm(int64(maxMemory))
}

func (r *request) Reader() (*multipart.Reader, error) {
	if r.limit >= 0 {
		r.r.Body = http.MaxBytesReader(r.w, r.r.Body, int64(r.limit))
	}
	return r.r.MultipartReader()
}

func (r *request) FormValues() url.Values {
	return r.r.Form
}

func (r *request) Done() <-chan struct{} {
	return r.r.Context().Done()
}

func (r *request) Form() *multipart.Form {
	return r.r.MultipartForm
}

func (r *request) FormValue(key string) string {
	return r.r.FormValue(key)
}

func (r *request) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return r.r.FormFile(key)
}

func (r *request) ResponseWriter() http.ResponseWriter {
	return r.w
}

func (r *request) RequestHeader() http.Header {
	return r.r.Header
}

func (r *request) ResponseHeader() http.Header {
	return r.w.Header()
}

type eventRequest[E any] struct {
	request
	e *E
}

func (e *eventRequest[E]) Event() E {
	return *e.e
}

type formHookRequest[D any] struct {
	request
	data *D
}

func (d *formHookRequest[D]) Data() D {
	return *d.data
}
