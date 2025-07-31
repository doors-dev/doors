package doors

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/doors-dev/doors/internal/front"
)

type afterFunc func(context.Context) (*front.After, error)

func (a afterFunc) after(ctx context.Context) (*front.After, error) {
	return a(ctx)
}

type After interface {
	after(context.Context) (*front.After, error)
}

func AfterLocationReload() After {
	return afterFunc(func(context.Context) (*front.After, error) {
		return &front.After{
			Name: "location_reload",
		}, nil
	})
}

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

type RAfter interface {
	After(After) error
}

// R provides common cookie-related methods for all request types.
type R interface {
	// SetCookie sets an HTTP cookie in the response.
	SetCookie(cookie *http.Cookie)

	// GetCookie retrieves a cookie from the request by name.
	GetCookie(name string) (*http.Cookie, error)
}

// REvent wraps a DOM event sent from the frontend.
//
// It includes basic request handling via BaseRequest and provides access
// to the decoded event payload of type E.
type REvent[E any] interface {
	R

	// Event returns the event payload.
	Event() E
	RAfter
}

// RForm provides access to decoded form data sent from the client.
//
// It includes cookie support and exposes structured form input via the Data method.
type RForm[D any] interface {
	R

	// Data returns the parsed form data as a typed value.
	Data() D
	RAfter
}

// RRawForm gives access to low-level multipart form data parsing.
//
// It is used when custom handling of form or file inputs is required.
type RRawForm interface {
	R

	// Reader returns a multipart.Reader for streaming multipart form data.
	Reader() (*multipart.Reader, error)

	// ParseForm parses the request form data into a ParsedForm.
	// maxMemory controls how much memory is used for non-file parts.
	ParseForm(maxMemory int) (ParsedForm, error)
	RAfter
}

// ParsedForm provides access to the contents of a parsed multipart form.
type ParsedForm interface {
	// FormValues returns the parsed URL-encoded form values.
	FormValues() url.Values

	// FormValue returns the first value for the given key.
	FormValue(key string) string

	// FormFile returns the uploaded file associated with the given key.
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)

	// Form returns the full multipart.Form object.
	Form() *multipart.Form
}

type RCall interface {
	RRawForm
	W() http.ResponseWriter
	Body() io.ReadCloser
	RAfter
}

type RHook[D any] interface {
	R
	Data() D
	RAfter
}

type RRawHook interface {
	RRawForm
	W() http.ResponseWriter
	Body() io.ReadCloser
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
