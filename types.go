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
	Group(string, ...*docs.PathDoc) IRouter
	OnError(handler interface{}) IRoutes
}

type IRoutes interface {
	Use(Middleware) IRoutes

	Handle(string, string, interface{}, ...*docs.PathDoc) IRoutes
	Any(string, interface{}, ...*docs.PathDoc) IRoutes
	GET(string, interface{}, ...*docs.PathDoc) IRoutes
	POST(string, interface{}, ...*docs.PathDoc) IRoutes
	DELETE(string, interface{}, ...*docs.PathDoc) IRoutes
	PATCH(string, interface{}, ...*docs.PathDoc) IRoutes
	PUT(string, interface{}, ...*docs.PathDoc) IRoutes
	OPTIONS(string, interface{}, ...*docs.PathDoc) IRoutes
	HEAD(string, interface{}, ...*docs.PathDoc) IRoutes
}

type Status int

type HeadersInterface interface {
	HeadersDocs()
}
type Headers http.Header

func (m Headers) HeadersDocs() {}

type QueryInterface interface {
	QueryDocs()
}
type Query struct{}

func (m Query) QueryDocs() {}

type BodyInterface interface {
	BodyDocs()
}
type Body struct{}

func (m Body) BodyDocs() {}

type ErrorInterface interface {
	ErrorDocs()
}
type ErrorResponse struct{}

func (m ErrorResponse) ErrorDocs() {}

type ResponseInterface interface {
	ResponseDocs()
}
type Response struct{}

func (m Response) ResponseDocs() {}

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

type MiddlewareFactory func() Middleware

type DefaultErrorResponse struct {
	ErrorResponse `status_codes:"400,422"`
	Message       string `json:"message"`
	Success       bool   `json:"success"`
}

func errorHandler(err error) (Status, *DefaultErrorResponse) {
	e := &DefaultErrorResponse{
		Success: false,
		Message: err.Error(),
	}

	switch err.(type) {
	default:
		return http.StatusInternalServerError, e
	}
}
