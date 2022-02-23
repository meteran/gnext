package gnext

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext/docs"
	"net/http"
)

func Router(docsOptions ...*docs.Options) *RootRouter {
	r := gin.Default()

	docsOptions = append(docsOptions, &docs.Options{})

	return &RootRouter{
		routerGroup: routerGroup{
			pathPrefix:   "",
			rawRouter:    r,
			middlewares:  nil,
			Docs:         docs.New(docsOptions[0]),
			errorHandler: DefaultErrorHandler,
		},
		engine: r,
	}
}

type RootRouter struct {
	routerGroup
	engine *gin.Engine
}

func (r *RootRouter) Engine() http.Handler {
	return r.engine
}

func (r *RootRouter) Run(host string, port string) error {
	r.Docs.RegisterRoutes(r.rawRouter, port)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r.engine,
	}

	return srv.ListenAndServe()
}
