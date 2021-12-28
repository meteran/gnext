package gnext

import (
	"github.com/gin-gonic/gin"
	"gnext.io/gnext/docs"
	"net/http"
)

type routerGroup struct {
	pathPrefix  string
	rawRouter   gin.IRouter
	middlewares []Middleware
	Docs        *docs.Docs
}

func (r *routerGroup) Any(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		r.Handle(method, path, handler, doc...)
	}
	return r
}

func (r *routerGroup) DELETE(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodDelete, path, handler, doc...)
}

func (r *routerGroup) PATCH(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodPatch, path, handler, doc...)
}

func (r *routerGroup) PUT(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodPut, path, handler, doc...)
}

func (r *routerGroup) OPTIONS(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodOptions, path, handler, doc...)
}

func (r *routerGroup) HEAD(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodHead, path, handler, doc...)
}

func (r *routerGroup) GET(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodGet, path, handler, doc...)
}

func (r *routerGroup) POST(path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	return r.Handle(http.MethodPost, path, handler, doc...)
}

func (r *routerGroup) Handle(method string, path string, handler interface{}, doc ...*docs.PathDoc) IRoutes {
	wrapper := WrapHandler(method, path, r.middlewares, r.Docs, handler, doc...)
	r.rawRouter.Handle(method, path, wrapper.rawHandle)
	return r
}

func (r *routerGroup) Use(middleware Middleware) IRoutes {
	r.middlewares = append(r.middlewares, middleware)
	return r
}

func (r *routerGroup) Group(prefix string, _ ...*docs.PathDoc) IRouter {
	return &routerGroup{
		pathPrefix:  prefix,
		rawRouter:   r.rawRouter.Group(prefix),
		middlewares: append([]Middleware{}, r.middlewares...),
		Docs:        r.Docs,
	}
}

func (r *routerGroup) RawRouter() gin.IRouter {
	return r.rawRouter
}
