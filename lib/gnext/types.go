package gnext

import "reflect"

type QueryInterface interface {
	QueryDocs()
}

type BodyInterface interface {
	BodyDocs()
}

type MarkQuery struct {}
func (m MarkQuery) QueryDocs() {}

type MarkBody struct {}
func (m MarkBody) BodyDocs() {}


var (
	queryType = reflect.TypeOf((*QueryInterface)(nil)).Elem()
	bodyType = reflect.TypeOf((*BodyInterface)(nil)).Elem()
)