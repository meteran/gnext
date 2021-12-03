package gnext

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func New() *Router {
	r := gin.Default()

	docs := NewDocs()
	r.GET("/docs", docs.handler)
	return &Router{
		engine: r,
		docs: docs,
	}
}

type Router struct {
	engine *gin.Engine
	docs   *Docs
}

func (r *Router) GET(path string, handler interface{}) {
	r.Handle(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler interface{}) {
	r.Handle(http.MethodPost, path, handler)
}


func (r *Router) Handle(method string, path string, handler interface{}) {
	wrapper := WrapHandler(method, path, handler)
	r.docs.append(wrapper.docs)
	r.engine.Handle(method, path, wrapper.rawHandle)
}

func (r *Router) Engine() http.Handler {
	return r.engine
}

