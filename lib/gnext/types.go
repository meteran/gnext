package gnext

import (
	"net/http"
	"reflect"
)

const (
	BodyParam  = "body"
	QueryParam = "query"
)

type Headers http.Header

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

type Error struct{}

func (m Error) ErrorDocs() {}

type ResponseInterface interface {
	ErrorDocs()
}

type Response struct{}

func (m Error) ResponseDocs() {}

var (
	queryType    = reflect.TypeOf((*QueryInterface)(nil)).Elem()
	bodyType     = reflect.TypeOf((*BodyInterface)(nil)).Elem()
	errorType    = reflect.TypeOf((*ErrorInterface)(nil)).Elem()
	responseType = reflect.TypeOf((*ErrorInterface)(nil)).Elem()
	headersType  = reflect.TypeOf(Headers{})
)

type Middleware struct {
	Before  interface{}
	After   interface{}
}

type MiddlewareFactory func() Middleware
