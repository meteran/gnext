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
	r.engine.GET(path, r.wrap(handler))
}

func (r *Router) POST(path string, handler interface{}) {
	r.engine.POST(path, r.wrap(handler))
}


func (r *Router) wrap(handler interface{}) gin.HandlerFunc {
	wrapper := WrapHandler(handler)
	r.docs.append(wrapper.docs)
	return wrapper.rawHandle
}

func (r *Router) Engine() http.Handler {
	return r.engine
}

