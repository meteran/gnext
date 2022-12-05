package gnext

import (
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext/docs"
	"net/http"
)

type routerGroup struct {
	pathPrefix    string
	rawRouter     gin.IRouter
	middlewares   middlewares
	Docs          *docs.Docs
	errorHandlers errorHandlers
}

func (g *routerGroup) OnError(errorHandler interface{}) IRoutes {
	g.errorHandlers.setup(errorHandler)
	return g
}

func (g *routerGroup) Any(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		g.Handle(method, path, handler, doc...)
	}
	return g
}

func (g *routerGroup) DELETE(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodDelete, path, handler, doc...)
}

func (g *routerGroup) PATCH(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodPatch, path, handler, doc...)
}

func (g *routerGroup) PUT(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodPut, path, handler, doc...)
}

func (g *routerGroup) OPTIONS(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodOptions, path, handler, doc...)
}

func (g *routerGroup) HEAD(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodHead, path, handler, doc...)
}

func (g *routerGroup) GET(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodGet, path, handler, doc...)
}

func (g *routerGroup) POST(path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	return g.Handle(http.MethodPost, path, handler, doc...)
}

func (g *routerGroup) Handle(method string, path string, handler interface{}, doc ...*docs.Endpoint) IRoutes {
	wrapper := WrapHandler(method, g.fullPath(path), g.middlewares, g.Docs, handler, g.errorHandlers, doc...)
	g.rawRouter.Handle(method, path, wrapper.requestHandler)
	return g
}

func (g *routerGroup) Use(middleware Middleware) IRoutes {
	g.middlewares = append(g.middlewares, &middleware)
	return g
}

func (g *routerGroup) Group(prefix string, _ ...*docs.Endpoint) IRouter {
	return &routerGroup{
		pathPrefix:    g.fullPath(prefix),
		rawRouter:     g.rawRouter.Group(prefix),
		middlewares:   g.middlewares.copy(),
		Docs:          g.Docs,
		errorHandlers: g.errorHandlers.copy(),
	}
}

func (g *routerGroup) RawRouter() gin.IRouter {
	return g.rawRouter
}

func (g *routerGroup) fullPath(path string) string {
	return g.pathPrefix + path
}
