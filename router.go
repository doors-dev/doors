package doors

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/router"
)

func Include() templ.Component {
	return front.Include
}

type Mod = router.Mod

type Page[M any] interface {
	Render(SourceBeam[M]) templ.Component
}

type PageRoute = router.Response

type PageRequest[M any] interface {
	BaseRequest
	GetModel() M
	RequestHeader() http.Header
	ResponseHeader() http.Header
}

type PageRouter[M any] interface {
	Page(page Page[M]) PageRoute
	PageStatus(page Page[M], status int) PageRoute
	PageFunc(pageFunc func(SourceBeam[M]) templ.Component) PageRoute
	PageFuncStatus(pageFunc func(SourceBeam[M]) templ.Component, status int) PageRoute
	Reroute(model any, detached bool) PageRoute
	Redirect(model any) PageRoute
	RedirectStatus(model any, status int) PageRoute
}

type pageRequest[M any] struct {
	r *router.Request[M]
}

func (r *pageRequest[M]) GetModel() M {
	return *r.r.Model
}

func (r *pageRequest[M]) RequestHeader() http.Header {
	return r.r.R.Header
}
func (r *pageRequest[M]) ResponseHeader() http.Header {
	return r.r.W.Header()
}
func (r *pageRequest[M]) GetCookie(name string) (*http.Cookie, error) {
	return r.r.R.Cookie(name)
}

func (r *pageRequest[M]) SetCookie(name string, cookie *http.Cookie) {
	http.SetCookie(r.r.W, cookie)
}

func (r *pageRequest[M]) Reroute(model any, detached bool) PageRoute {
	return &router.RerouteResponse{
		Detached: detached,
		Model:    model,
	}
}

func (r *pageRequest[M]) Redirect(model any) PageRoute {
	return &router.RedirectResponse{
		Model: model,
	}
}

func (r *pageRequest[M]) RedirectStatus(model any, status int) PageRoute {
	return &router.RedirectResponse{
		Model:  model,
		Status: status,
	}
}

func (r *pageRequest[M]) Page(page Page[M]) PageRoute {
	return r.PageStatus(page, 200)
}

func (r *pageRequest[M]) PageStatus(page Page[M], status int) PageRoute {
	return &router.PageResponse[M]{
		Page:    page,
		Model:   r.r.Model,
		Adapter: r.r.Adapter,
		Status:  status,
	}
}
func (r *pageRequest[M]) StaticPageStatus(content templ.Component, status int) PageRoute {
	return &router.StaticPage{
		Content: content,
		Status:  status,
	}
}
func (r *pageRequest[M]) StaticPage(content templ.Component) PageRoute {
	return r.StaticPageStatus(content, 200)
}

type pageFunc[M any] func(SourceBeam[M]) templ.Component

func (p pageFunc[M]) Render(b SourceBeam[M]) templ.Component {
	return p(b)
}

func (r *pageRequest[M]) PageFunc(f func(SourceBeam[M]) templ.Component) PageRoute {
	return r.PageFuncStatus(f, 200)
}

func (r *pageRequest[M]) PageFuncStatus(f func(SourceBeam[M]) templ.Component, status int) PageRoute {
	return r.PageStatus(pageFunc[M](f), status)
}

func ServePage[M any](handler func(PageRouter[M], PageRequest[M]) PageRoute) Mod {
	return router.RoutePage(func(r *router.Request[M]) router.Response {
		pr := &pageRequest[M]{
			r: r,
		}
		return handler(pr, pr)
	})
}

func ServeDirPath(prefix string, localPath string) Mod {
	return router.ServeDirPath(prefix, localPath)
}

func ServeDir(prefix string, root http.FileSystem) Mod {
	return router.ServeDir(prefix, root)
}

func ServeFile(path string, localPath string) Mod {
	return router.ServeFile(path, localPath)
}

func ServeRaw(path string, handler func(w http.ResponseWriter, r *http.Request)) Mod {
	return router.ServeRaw(path, handler)
}

type Router interface {
	http.Handler
	Use(...Mod)
}

func NewRouter() Router {
	return router.NewRouter()
}

func ServeFallback(handler http.Handler) Mod {
	return router.ServeFallback(handler)
}

func SetGoroutineLimitPerInstance(n int) Mod {
	return router.SetGoroutineLimit(n)
}

func SetSessionHooks(create func(id string), delete func(id string)) Mod {
	return router.SetSessionHooks(create, delete)
}

type SystemConf = common.SystemConf

func SetSystemConf(conf SystemConf) Mod {
	return router.SetSystemConf(conf)
}

func SetErrorPage(page func(message string) templ.Component) Mod {
	return router.SetErrorPage(page)
}

type CSP = common.CSP

func SetCSP(csp *CSP) Mod {
	return router.SetCSP(csp)
}
