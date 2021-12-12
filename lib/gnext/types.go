package gnext

import (
	"net/http"
	"reflect"
)

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
type Error struct{}
func (m Error) ErrorDocs() {}

type ResponseInterface interface {
	ResponseDocs()
}
type Response struct{}
func (m Response) ResponseDocs() {}

var (
	queryType    = reflect.TypeOf((*QueryInterface)(nil)).Elem()
	bodyType     = reflect.TypeOf((*BodyInterface)(nil)).Elem()
	errorType    = reflect.TypeOf((*error)(nil)).Elem()
	responseType = reflect.TypeOf((*ResponseInterface)(nil)).Elem()
	headersType  = reflect.TypeOf((*HeadersInterface)(nil)).Elem()
	stringType   = reflect.TypeOf("")
	intType      = reflect.TypeOf(0)
)

type Middleware struct {
	Before interface{}
	After  interface{}
}

type MiddlewareFactory func() Middleware
