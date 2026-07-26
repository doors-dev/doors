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

// RequestAfter is the part of a request handle that runs client-side actions
// after a successful response.
type RequestAfter interface {
	// After registers actions to run on the client once the response
	// succeeds. A second call replaces the actions of the first.
	After(a Actions) error
}

// RequestCommon is the HTTP-level part of every request handle Doors passes to
// a handler.
type RequestCommon interface {
	// SetCookie adds a cookie to the response.
	SetCookie(cookie *http.Cookie)
	// GetCookie returns the named request cookie, or [http.ErrNoCookie] if it
	// is not present.
	GetCookie(name string) (*http.Cookie, error)
	// Context returns the context of the underlying HTTP request. It is not
	// the Doors runtime context the handler receives.
	Context() context.Context
	// RequestHeader returns the incoming request headers.
	RequestHeader() http.Header
	// RemoteAddr returns the address of the direct peer as host:port. Proxy
	// headers are not applied.
	RemoteAddr() string
}

// RequestEvent is the request handle passed to event attr handlers.
type RequestEvent[E any] interface {
	RequestCommon
	RequestAfter
	// Event returns the decoded event payload.
	Event() E
}

// RequestForm is the request handle passed to [ASubmit] handlers.
type RequestForm[D any] interface {
	RequestCommon
	RequestAfter
	// Data returns the submitted form decoded into D.
	Data() D
}

// RequestRawForm is the request handle passed to [ARawSubmit] handlers, with
// the multipart body left undecoded.
type RequestRawForm interface {
	RequestCommon
	RequestAfter
	// SetRequestBodyLimit sets the max body size in bytes accepted by later
	// Body, ParseForm, and Reader calls. A negative limit disables the
	// check; zero rejects a non-empty body. Default: ServerRequestBodyLimit
	// of [Conf].
	SetRequestBodyLimit(limit int)
	// ResponseWriter returns the writer for the response.
	ResponseWriter() http.ResponseWriter
	// Reader returns a multipart reader for streaming the form one part at a
	// time. Unlike [RequestRawForm.ParseForm], it buffers no part to disk.
	Reader() (*multipart.Reader, error)
	// ParseForm parses the whole multipart body, keeping up to maxMemory
	// bytes of file parts in memory and storing the rest in temporary files.
	// It fails if the body is not a multipart form.
	ParseForm(maxMemory int) (ParsedForm, error)
}

// ParsedForm is a parsed multipart form.
type ParsedForm interface {
	// FormValues returns the form fields merged with the URL query values.
	FormValues() url.Values
	// FormValue returns the first value for key, empty if there is none.
	FormValue(key string) string
	// FormFile returns the first uploaded file for key, or an error if there
	// is none.
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	// Form returns the parsed values and file headers.
	Form() *multipart.Form
}

// RequestHook is the request handle passed to [AHook] handlers.
type RequestHook[D any] interface {
	RequestCommon
	RequestAfter
	// Data returns the hook argument decoded from JSON into D.
	Data() D
}

// RequestRawHook is the request handle passed to [ARawHook] handlers, with the
// body left undecoded.
type RequestRawHook interface {
	RequestRawForm
	// Body returns the request body, bounded by the limit from
	// [RequestRawForm.SetRequestBodyLimit].
	Body() io.ReadCloser
}

// Request is the request handle passed to the page factory, for cookies and
// request or response headers.
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
