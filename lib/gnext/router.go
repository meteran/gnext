package gnext

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gnext.io/gnext/docs"
	"net/http"
)

func New(documentation *docs.Docs) *Router {
	r := gin.Default()

	documentation.NewOpenAPI()

	err := documentation.Valid()
	if err != nil {
		panic(err)
	}

	docHandler := docs.NewHandler(documentation)
	r.LoadHTMLGlob("lib/gnext/templates/*.html")

	docGroup := r.Group(documentation.OpenAPIPath)
	docGroup.GET("", docHandler.Docs)
	docGroup.GET("/openapi.json", docHandler.File)
	docGroup.Use(cors.New(*documentation.CORSConfig()))


	return &Router{
		engine:        r,
		documentation: documentation,
	}
}

type Router struct {
	engine        *gin.Engine
	documentation *docs.Docs
	middlewares   []Middleware
}

func (r *Router) GET(path string, handler interface{}, doc *docs.PathDoc) {
	r.Handle(http.MethodGet, path, handler, doc)
}

func (r *Router) POST(path string, handler interface{}, doc ...*docs.PathDoc) {
	r.Handle(http.MethodPost, path, handler, doc...)
}

func (r *Router) Handle(method string, path string, handler interface{}, doc ...*docs.PathDoc) {
	wrapper := WrapHandler(method, path, r.middlewares, r.documentation, handler, doc...)
	r.engine.Handle(method, path, wrapper.rawHandle)
}

func (r *Router) Engine() http.Handler {
	return r.engine
}

func (r *Router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) Docs() *docs.Docs {
	return r.documentation
}

func (r *Router) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: r.Engine(),
	}
	if !r.Docs().InMemory {
		err := r.Docs().Build()
		if err != nil {
			panic(fmt.Sprintf("cannot build documentation; error: %v", err))
		}
	}

	return srv.ListenAndServe()
}
