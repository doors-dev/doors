// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"context"
	"github.com/doors-dev/doors/internal/front"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type afterFunc func(context.Context) (*front.After, error)

func (a afterFunc) after(ctx context.Context) (*front.After, error) {
	return a(ctx)
}

// After represents a client-side action to execute after a request completes.
// After actions can perform navigation, page reloads, or other browser operations.
type After interface {
	after(context.Context) (*front.After, error)
}

// AfterLocationReload creates an After action that reloads the current page.
// This triggers a full page refresh in the browser.
func AfterLocationReload() After {
	return afterFunc(func(context.Context) (*front.After, error) {
		return &front.After{
			Name: "location_reload",
		}, nil
	})
}

// AfterLocationAssign creates an After action that navigates to a new URL.
// This adds the new URL to the browser's history stack. The model parameter
// is used to generate the target URL using the application's routing system.
func AfterScrollInto(selector string, smooth bool) After {
	return afterFunc(func(ctx context.Context) (*front.After, error) {
		return &front.After{
			Name: "scroll_into",
			Arg:  []any{selector, smooth},
		}, nil
	})
}

// AfterLocationAssign creates an After action that navigates to a new URL.
// This adds the new URL to the browser's history stack. The model parameter
// is used to generate the target URL using the application's routing system.
func AfterLocationAssign(model any) After {
	return afterFunc(func(ctx context.Context) (*front.After, error) {
		l, err := NewLocation(ctx, model)
		if err != nil {
			return nil, err
		}
		return &front.After{
			Name: "location_assign",
			Arg:  []any{l.String(), true},
		}, nil
	})
}

// AfterLocationReplace creates an After action that replaces the current URL.
// This navigates to a new URL without adding an entry to the browser's history stack.
// The model parameter is used to generate the target URL using the application's routing system.
func AfterLocationReplace(model any) After {
	return afterFunc(func(ctx context.Context) (*front.After, error) {
		l, err := NewLocation(ctx, model)
		if err != nil {
			return nil, err
		}
		return &front.After{
			Name: "location_replace",
			Arg:  []any{l.String(), true},
		}, nil
	})
}

// RAfter provides the ability to set an After action to execute client-side
// after the request completes.
type RAfter interface {
	After(After) error
}

// R provides basic request functionality including cookie management.
type R interface {
	SetCookie(cookie *http.Cookie)
	GetCookie(name string) (*http.Cookie, error)
	Done() <-chan struct{}
}

// REvent represents a request context for event handlers with typed event data.
// The generic type E represents the structure of the event data.
type REvent[E any] interface {
	R
	Event() E // Returns the event data
	RAfter
}

// RForm represents a request context for form submissions with typed form data.
// The generic type D represents the structure of the parsed form data.
type RForm[D any] interface {
	R
	Data() D // Returns the parsed form data
	RAfter
}

// RRawForm provides access to raw multipart form data and parsing capabilities.
// This is used when you need direct access to the form parsing process or
// when working with file uploads.
type RRawForm interface {
	R
	W() http.ResponseWriter
	Reader() (*multipart.Reader, error)          // Returns a multipart reader for streaming
	ParseForm(maxMemory int) (ParsedForm, error) // Parses form data with memory limit
	RAfter
}

// ParsedForm provides access to parsed form data including values and files.
type ParsedForm interface {
	FormValues() url.Values                                             // Returns all form values
	FormValue(key string) string                                        // Returns a single form value
	FormFile(key string) (multipart.File, *multipart.FileHeader, error) // Returns an uploaded file
	Form() *multipart.Form                                              // Returns the underlying multipart form
}

// RCall represents a request context for direct HTTP calls with full access
// to the response writer and request body. This provides the most control
// over the HTTP request/response cycle.
type RCall interface {
	RRawForm
	Body() io.ReadCloser // Returns the request body
	RAfter
}

// RHook represents a request context for hook handlers with typed data.
// The generic type D represents the structure of the hook data.
type RHook[D any] interface {
	R
	Data() D // Returns the hook data
	RAfter
}

// RRawHook provides access to raw request data for hook handlers without
// automatic data parsing. This gives full control over request processing.
type RRawHook interface {
	RRawForm
	Body() io.ReadCloser // Returns the request body
}

type request struct {
	w   http.ResponseWriter
	r   *http.Request
	ctx context.Context
}

func (r *request) After(a After) error {
	after, err := a.after(r.ctx)
	if err != nil {
		return err
	}
	err = after.Set(r.w.Header())
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
