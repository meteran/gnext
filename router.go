package gnext

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext/docs"
	"net/http"
)

func Router() *RootRouter {
	r := gin.Default()

	return &RootRouter{
		routerGroup: routerGroup{
			pathPrefix:   "",
			rawRouter:    r,
			middlewares:  nil,
			Docs:         nil,
			errorHandler: DefaultErrorHandler,
		},
		engine: r,
	}
}

func DocumentedRouter(documentation *docs.Docs) *RootRouter {
	r := gin.Default()

	documentation.NewOpenAPI()

	err := documentation.Valid()
	if err != nil {
		panic(err)
	}

	docHandler := docs.NewHandler(documentation)
	docGroup := r.Group(documentation.OpenAPIPath)

	docGroup.GET("", docHandler.Docs)
	docGroup.GET("/openapi.json", docHandler.File)

	return &RootRouter{
		routerGroup: routerGroup{
			pathPrefix:   "",
			rawRouter:    r,
			middlewares:  nil,
			Docs:         documentation,
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

func (r *RootRouter) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: r.engine,
	}
	if r.Docs != nil && !r.Docs.InMemory {
		err := r.Docs.Build()
		if err != nil {
			panic(fmt.Sprintf("cannot build documentation; error: %v", err))
		}
	}

	return srv.ListenAndServe()
}
