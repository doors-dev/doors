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

	"github.com/doors-dev/doors/internal/ctex"
)

// RequestAfter schedules client-side actions to run after a successful
// request.
type RequestAfter interface {
	// After appends client-side actions to the successful response.
	After([]Action) error
}

// Request exposes the common server-side request helpers available to Doors
// handlers.
type Request interface {
	// SetCookie adds a cookie to the response.
	SetCookie(cookie *http.Cookie)
	// GetCookie retrieves a cookie by name.
	GetCookie(name string) (*http.Cookie, error)
}

// RequestEvent is the request context passed to event handlers.
type RequestEvent[E any] interface {
	Request
	RequestAfter
	// Event returns the event payload.
	Event() E
}

// RequestForm is the request context passed to decoded form handlers.
type RequestForm[D any] interface {
	Request
	RequestAfter
	// Data returns the parsed form payload.
	Data() D
}

// RequestRawForm is the request context passed to raw multipart form handlers.
type RequestRawForm interface {
	Request
	RequestAfter
	// W returns the HTTP response writer.
	W() http.ResponseWriter
	// Reader returns a multipart reader for streaming form parts.
	Reader() (*multipart.Reader, error)
	// ParseForm parses the form data with a memory limit.
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
	Request
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

// RequestModel is the request context passed to [UseModel] handlers.
//
// Use it for cookies, request and response headers, and session-scoped state
// while deciding which [Response] to return.
type RequestModel interface {
	Request
	// SessionStore returns session-scoped storage.
	SessionStore() Store
	// RequestHeader returns the incoming request headers.
	RequestHeader() http.Header
	// ResponseHeader returns the outgoing response headers.
	ResponseHeader() http.Header
}

type request struct {
	w   http.ResponseWriter
	r   *http.Request
	ctx context.Context
}

func (r *request) After(action []Action) error {
	actions := intoActions(r.ctx, action)
	err := actions.Set(r.w.Header())
	if err != nil {
		panic(err)
	}
	return nil
}

func (r *request) Body() io.ReadCloser {
	return r.r.Body
}

func (r *request) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.w, cookie)
}

func (r *request) GetCookie(name string) (*http.Cookie, error) {
	return r.r.Cookie(name)
}

func (r *request) ParseForm(maxMemory int) (ParsedForm, error) {
	if maxMemory <= 0 {
		maxMemory = defaultMaxMemory
	}
	return r, r.r.ParseMultipartForm(int64(maxMemory))
}

func (r *request) Reader() (*multipart.Reader, error) {
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

func (r *request) W() http.ResponseWriter {
	return r.w
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

type modelRequest struct {
	request
	store ctex.Store
}

func (r *modelRequest) RequestHeader() http.Header {
	return r.request.r.Header
}

func (r *modelRequest) ResponseHeader() http.Header {
	return r.request.w.Header()
}

func (r *modelRequest) SessionStore() Store {
	return r.store
}
