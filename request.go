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
)

// RAfter allows setting client-side actions to run after a request completes.
type RAfter interface {
	// After sets client-side actions to run once the request finishes.
	After([]Action) error
}

// R provides basic request operations including cookie management.
type R interface {
	// SetCookie adds a cookie to the response.
	SetCookie(cookie *http.Cookie)
	// GetCookie retrieves a cookie by name.
	GetCookie(name string) (*http.Cookie, error)
	// Done signals when the request context is canceled or completed.
	Done() <-chan struct{}
}

// REvent provides request handling for event hooks with typed event data.
type REvent[E any] interface {
	R
	RAfter
	// Event returns the event payload.
	Event() E
}

// RForm provides request handling for form submissions with typed form data.
type RForm[D any] interface {
	R
	RAfter
	// Data returns the parsed form payload.
	Data() D
}

// RRawForm provides access to raw multipart form data for streaming or custom parsing.
type RRawForm interface {
	R
	RAfter
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

// RHook provides request handling for hook handlers with typed data.
type RHook[D any] interface {
	R
	RAfter
	// Data returns the parsed hook payload.
	Data() D
}

// RRawHook provides access to raw request data for hook handlers without parsing.
type RRawHook interface {
	RRawForm
	// Body returns the raw request body reader.
	Body() io.ReadCloser
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
