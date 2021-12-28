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
	docGroup.Use(cors.New(*documentation.CORSConfig()))
	docGroup.GET("", docHandler.Docs)
	docGroup.GET("/openapi.json", docHandler.File)

	return &Router{
		routerGroup: routerGroup{
			pathPrefix:  "",
			rawRouter:   r,
			middlewares: nil,
			Docs:        documentation,
		},
		engine: r,
	}
}

type Router struct {
	routerGroup
	engine        *gin.Engine
	documentation *docs.Docs
}

func (r *Router) Engine() http.Handler {
	return r.engine
}

func (r *Router) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: r.engine,
	}
	if !r.Docs.InMemory {
		err := r.Docs.Build()
		if err != nil {
			panic(fmt.Sprintf("cannot build documentation; error: %v", err))
		}
	}

	return srv.ListenAndServe()
}
