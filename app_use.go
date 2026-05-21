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

// UseResource serves a static Doors resource at path.
//
// path must not be "/" or empty. contentType is passed to the resource
// registry and should describe the served resource, for example "text/css".
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

// UseFS serves files from fsys under prefix.
//
// prefix must not be "/" or empty. cacheControl is written on successful file
// responses when non-empty.
func UseFS(prefix string, fsys fs.FS, cacheControl string) Use {
	if prefix == "/" || prefix == "" {
		panic(errors.New("ServeFS cannot serve the root prefix"))
	}
	return serveFileSystem(prefix, http.FS(fsys), cacheControl)
}

// UseDir serves files from a local directory under prefix.
//
// prefix must not be "/" or empty. cacheControl is written on successful file
// responses when non-empty.
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

// UseFile serves one local file at path.
//
// path must not be "/" or empty. cacheControl is written on successful file
// responses when non-empty.
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

// Cache-Control header presets for common use cases.
const (
	// CacheControlImmutable is for fingerprinted/hashed static assets
	// (e.g. app.a3f9b2.js, fonts, versioned images) that never change
	// at a given URL. Cached for 1 year, no revalidation.
	CacheControlImmutable = "public, max-age=31536000, immutable"

	// CacheControlStatic is for non-fingerprinted static files
	// (e.g. /favicon.ico, /robots.txt). Cached for 1 hour, revalidated after.
	CacheControlStatic = "public, max-age=3600, must-revalidate"

	// CacheControlStaticShort is for static files that may change occasionally
	// and where staleness is tolerable for a few minutes.
	CacheControlStaticShort = "public, max-age=300, must-revalidate"

	// CacheControlHTML is for HTML entry points (index.html, SSR pages)
	// where updates should be picked up quickly. Pair with an ETag.
	CacheControlHTML = "public, max-age=0, must-revalidate"

	// CacheControlCDN splits browser TTL (1h) from shared/CDN TTL (1d).
	// Useful when you want fast invalidation at the edge.
	CacheControlCDN = "public, max-age=3600, s-maxage=86400"

	// CacheControlPrivate is for per-user responses (dashboards, account pages).
	// Browsers may cache briefly, CDNs/proxies must not.
	CacheControlPrivate = "private, max-age=0, must-revalidate"

	// CacheControlNoCache forces revalidation on every request, but allows
	// storage. Use when content changes often but ETag/Last-Modified can
	// produce cheap 304s.
	CacheControlNoCache = "no-cache"

	// CacheControlNoStore disables caching entirely. Use for sensitive
	// responses (auth tokens, personal data, payment flows).
	CacheControlNoStore = "no-store"

	// CacheControlAPI is a sensible default for JSON API responses that
	// shouldn't be cached by intermediaries but can be revalidated.
	CacheControlAPI = "private, no-cache"
)
