package doors

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
)

type responseWriter struct {
	w           http.ResponseWriter
	setHeaders  func(h http.Header)
	wroteHeader bool
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	if code >= 200 && code < 300 {
		w.setHeaders(w.w.Header())
	}
	w.w.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.setHeaders(w.w.Header())
	}
	return w.w.Write(b)
}

func normalizePrefix(prefix string) string {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return prefix
}

func serveFS(prefix string, fsys http.FileSystem, cacheControl string, w http.ResponseWriter, r *http.Request) {
	rw := &responseWriter{
		w: w,
		setHeaders: func(h http.Header) {
			if cacheControl == "" {
				return
			}
			h.Set("Cache-Control", cacheControl)
		},
	}
	http.StripPrefix(prefix, http.FileServer(fsys)).ServeHTTP(rw, r)
}

// UseResource serves resource as a managed resource at path.
//
// path is matched exactly, with a leading slash added if it lacks one, and must
// be neither empty nor "/". contentType is sent as the Content-Type. Requests
// that do not match, and non-GET requests, go to the next handler. UseResource
// panics on an unusable path or resource.
//
// Responses carry the ServerCacheControl value from [Conf] and are gzipped
// unless ServerDisableGzip is set.
func UseResource(path string, resource ResourceStatic, contentType string) Use {
	if path == "/" || path == "" {
		panic(errors.New("ServeResource cannot serve the root path"))
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if resource == nil {
		panic(errors.New("ServeResource requires a static resource"))
	}
	entry := resource.StaticEntry()
	if entry == nil {
		panic(errors.New("ServeResource returned a nil static entry"))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.URL.Path != path {
				next.ServeHTTP(w, r)
				return
			}
			sess := r.Context().Value(common.KeySession).(instance.Session)
			res, err := sess.App().ResourceRegistry().Static(entry, contentType)
			if err != nil {
				sess.Logger().Error("ServeResource failed to prepare the resource", "error", err)
				w.WriteHeader(500)
				return
			}
			res.Serve(w, r)
		})
	}
}

// UseFS serves fsys under prefix.
//
// prefix must be neither empty nor "/", and is normalized to have a leading and
// trailing slash. cacheControl is set on successful responses; empty leaves the
// header alone. Requests outside prefix, and non-GET requests, go to the next
// handler.
func UseFS(prefix string, fsys fs.FS, cacheControl string) Use {
	if prefix == "/" || prefix == "" {
		panic(errors.New("ServeFS cannot serve the root prefix"))
	}
	return serveFileSystem(prefix, http.FS(fsys), cacheControl)
}

// UseDir serves the local directory dirPath under prefix.
//
// prefix must be neither empty nor "/", and is normalized to have a leading and
// trailing slash. cacheControl is set on successful responses; empty leaves the
// header alone. Requests outside prefix, and non-GET requests, go to the next
// handler.
func UseDir(prefix string, dirPath string, cacheControl string) Use {
	if prefix == "/" || prefix == "" {
		panic(errors.New("ServeDir cannot serve the root prefix"))
	}
	return serveFileSystem(prefix, http.Dir(dirPath), cacheControl)
}

func serveFileSystem(prefix string, fsys http.FileSystem, cacheControl string) Use {
	prefix = normalizePrefix(prefix)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
			serveFS(prefix, fsys, cacheControl, w, r)
		})
	}
}

// UseFile serves the local file filePath at path.
//
// path is matched exactly, with a leading slash added if it lacks one, and must
// be neither empty nor "/". cacheControl is set on successful responses; empty
// leaves the header alone. Requests that do not match, and non-GET requests, go
// to the next handler.
func UseFile(path string, filePath string, cacheControl string) Use {
	if path == "/" || path == "" {
		panic(errors.New("ServeFile cannot serve the root path"))
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.URL.Path != path {
				next.ServeHTTP(w, r)
				return
			}
			rw := &responseWriter{
				w: w,
				setHeaders: func(h http.Header) {
					if cacheControl == "" {
						return
					}
					h.Set("Cache-Control", cacheControl)
				},
			}
			http.ServeFile(rw, r, filePath)
		})
	}
}

// Cache-Control presets for [UseFS], [UseDir], and [UseFile].
const (
	// CacheControlImmutable caches for a year and never revalidates. Use for
	// hash-named URLs.
	CacheControlImmutable = "public, max-age=31536000, immutable"

	// CacheControlStatic caches for an hour, then revalidates.
	CacheControlStatic = "public, max-age=3600, must-revalidate"

	// CacheControlStaticShort caches for five minutes, then revalidates.
	CacheControlStaticShort = "public, max-age=300, must-revalidate"

	// CacheControlHTML lets shared caches store the response but always
	// revalidates it.
	CacheControlHTML = "public, max-age=0, must-revalidate"

	// CacheControlCDN caches for an hour in browsers, a day in shared caches.
	CacheControlCDN = "public, max-age=3600, s-maxage=86400"

	// CacheControlPrivate keeps the response out of shared caches and always
	// revalidates it.
	CacheControlPrivate = "private, max-age=0, must-revalidate"

	// CacheControlNoCache stores the response but always revalidates it.
	CacheControlNoCache = "no-cache"

	// CacheControlNoStore forbids storing the response anywhere.
	CacheControlNoStore = "no-store"

	// CacheControlAPI is CacheControlNoCache restricted to private caches.
	CacheControlAPI = "private, no-cache"
)
