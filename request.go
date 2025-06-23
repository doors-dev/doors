package doors

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// BaseRequest provides common cookie-related methods for all request types.
type BaseRequest interface {
	// SetCookie sets an HTTP cookie in the response.
	SetCookie(name string, cookie *http.Cookie)

	// GetCookie retrieves a cookie from the request by name.
	GetCookie(name string) (*http.Cookie, error)
}

// EventRequest wraps a DOM event sent from the frontend.
//
// It includes basic request handling via BaseRequest and provides access
// to the decoded event payload of type E.
type EventRequest[E any] interface {
	BaseRequest

	// Event returns the event payload.
	Event() E
}

// FormRequest provides access to decoded form data sent from the client.
//
// It includes cookie support and exposes structured form input via the Data method.
type FormRequest[D any] interface {
	BaseRequest

	// Data returns the parsed form data as a typed value.
	Data() D
}


// RawFormRequest gives access to low-level multipart form data parsing.
//
// It is used when custom handling of form or file inputs is required.
type RawFormRequest interface {
	BaseRequest

	// Reader returns a multipart.Reader for streaming multipart form data.
	Reader() (*multipart.Reader, error)

	// ParseForm parses the request form data into a ParsedForm.
	// maxMemory controls how much memory is used for non-file parts.
	ParseForm(maxMemory int) (ParsedForm, error)
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


type CallRequest interface {
	RawFormRequest
	W() http.ResponseWriter
	Body() io.ReadCloser
}

type HookRequest[D any] interface {
    BaseRequest
    Data() D
}

type RawHookRequest interface {
	RawFormRequest
	W() http.ResponseWriter
	Body() io.ReadCloser
}

type request struct {
	w http.ResponseWriter
	r *http.Request
}

func (r *request) Body() io.ReadCloser {
	return r.r.Body
}
func (r *request) SetCookie(name string, cookie *http.Cookie) {
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
