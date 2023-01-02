package gnext

import (
	"github.com/gin-gonic/gin"
	"github.com/meteran/gnext/docs"
	"net/http"
	"reflect"
)

type IRouter interface {
	IRoutes
	RawRouter() gin.IRouter
	Group(string, ...*docs.Endpoint) IRouter
	OnError(handler interface{}) IRoutes
}

type IRoutes interface {
	Use(Middleware) IRoutes

	Handle(string, string, interface{}, ...*docs.Endpoint) IRoutes
	Any(string, interface{}, ...*docs.Endpoint) IRoutes
	GET(string, interface{}, ...*docs.Endpoint) IRoutes
	POST(string, interface{}, ...*docs.Endpoint) IRoutes
	DELETE(string, interface{}, ...*docs.Endpoint) IRoutes
	PATCH(string, interface{}, ...*docs.Endpoint) IRoutes
	PUT(string, interface{}, ...*docs.Endpoint) IRoutes
	OPTIONS(string, interface{}, ...*docs.Endpoint) IRoutes
	HEAD(string, interface{}, ...*docs.Endpoint) IRoutes
}

type Status int

type HeadersInterface interface {
	GnHeaders()
}
type Headers http.Header

func (m Headers) GnHeaders() {}

type QueryInterface interface {
	GnQuery()
}
type Query struct{}

func (m Query) GnQuery() {}

type BodyInterface interface {
	GnBody()
}
type Body struct{}

func (m Body) GnBody() {}

type ErrorResponse struct{}

type ResponseInterface interface {
	GnResponse()
}
type Response struct{}

func (m Response) GnResponse() {}

var (
	queryInterfaceType    = reflect.TypeOf((*QueryInterface)(nil)).Elem()
	bodyInterfaceType     = reflect.TypeOf((*BodyInterface)(nil)).Elem()
	errorInterfaceType    = reflect.TypeOf((*error)(nil)).Elem()
	responseInterfaceType = reflect.TypeOf((*ResponseInterface)(nil)).Elem()
	headersInterfaceType  = reflect.TypeOf((*HeadersInterface)(nil)).Elem()

	rawContextType = reflect.TypeOf(&gin.Context{})
	headersType    = reflect.TypeOf(Headers{})
	statusType     = reflect.TypeOf(Status(0))
	stringType     = reflect.TypeOf("")
	intType        = reflect.TypeOf(0)
)

type Middleware struct {
	Before interface{}
	After  interface{}
}

type middlewares []*Middleware

func (m middlewares) copy() middlewares {
	return append(middlewares{}, m...)
}

func (m middlewares) count() int {
	count := 0
	for _, middleware := range m {
		if middleware.Before != nil {
			count++
		}
		if middleware.After != nil {
			count++
		}
	}
	return count
}

type DefaultErrorResponse struct {
	ErrorResponse `default_status:"500" status_codes:"4XX,5XX"`
	Message       string   `json:"message"`
	Details       []string `json:"details"`
	Success       bool     `json:"success"`
}
