package gnext

import (
	"net/http"
	"reflect"
)

const (
	BodyParam = "body"
	QueryParam = "query"
)
type Headers http.Header

type QueryInterface interface {
	QueryDocs()
}

type BodyInterface interface {
	BodyDocs()
}

type Query struct {}
func (m Query) QueryDocs() {}

type Body struct {}
func (m Body) BodyDocs() {}


var (
	queryType = reflect.TypeOf((*QueryInterface)(nil)).Elem()
	bodyType = reflect.TypeOf((*BodyInterface)(nil)).Elem()
	headersType = reflect.TypeOf(Headers{})
)