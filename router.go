package gnext

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext/docs"
	"log"
	"net/http"
	"strings"
)

// Router is a RootRouter constructor. It gets one optional parameter *docs.Options.
// If passed, all non-empty fields from this struct will be used to initialize the documentation.
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

// RootRouter is a main struct of the module.
// All other operations are made using this router.
type RootRouter struct {
	routerGroup
	engine *gin.Engine
}

// Engine returns the raw Gin engine.
// It can be used to add Gin-native handlers or middlewares.
//
// Note: handlers and middleware attached directly to the raw engine, bypasses the gNext core. Because of that they won't be included in the docs nor validation mechanism.
func (r *RootRouter) Engine() *gin.Engine {
	return r.engine
}

// Run starts the http server. It takes optional address parameters. The number of parameters is meaningful:
//  * 0 - defaults to ":8080".
//  * 1 - means the given address is either a full address in form 'host:port` or, if doesn't contain ':',  a port.
//  * 2 - first parameter is a host, the latter one is a port.
//  * 3+ - invalid address.
func (r *RootRouter) Run(address ...string) error {
	host, port := resolveAddress(address)
	r.Docs.RegisterRoutes(r.rawRouter, port)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r.engine,
	}

	if host == "" {
		host = "localhost"
	}

	log.Printf("starting server on http://%s:%s", host, port)
	return srv.ListenAndServe()
}

func resolveAddress(address []string) (string, string) {
	host := ""
	port := "8080"

	switch len(address) {
	case 0:
		port = "8080"
	case 1:
		addr := strings.Split(address[0], ":")
		switch len(addr) {
		case 1:
			port = addr[0]
		case 2:
			host = addr[0]
			port = addr[1]
		default:
			panic(fmt.Sprintf("invalid address: '%s'", address[0]))
		}
	case 2:
		host = address[0]
		port = address[1]
	default:
		panic(fmt.Sprintf("invalid number of arguments: %d", len(address)))
	}
	return host, port
}
