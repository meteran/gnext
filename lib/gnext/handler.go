package gnext

import (
	"github.com/gin-gonic/gin"
	"reflect"
)

func WrapHandler(handler interface{}) *HandlerWrapper {
	wrapper := &HandlerWrapper{
		originalHandler: handler,
	}
	wrapper.init()
	return wrapper
}

type HandlerWrapper struct {
	originalHandler interface{}
	docs            interface{}
	handlerValue    reflect.Value
	queryType       reflect.Type
	bodyType        reflect.Type
}

func (w *HandlerWrapper) init() {
	w.docs = gin.H{"elo": "ole"}

	w.handlerValue = reflect.ValueOf(w.originalHandler)
	ht := reflect.TypeOf(w.originalHandler)
	if ht.Kind() != reflect.Func {
		panic("handler is not a function")
	}

	for i := 0; i < ht.NumIn(); i++ {
		inParam := ht.In(i)
		if inParam.Kind() == reflect.Ptr {
			inParam = inParam.Elem()
		}

		println(inParam.Name())
		println(inParam.Implements(bodyType))
		println(inParam.Implements(queryType))
		//
		//if inParam.Kind() == reflect.Struct {
		//	w.queryOrBody(inParam)
		//}
	}

	ht.NumOut()
}

func (w *HandlerWrapper) queryOrBody(inParam reflect.Type) {
	for k := 0; k < inParam.NumField(); k++ {
		field := inParam.Field(k)
		if field.Anonymous {
			switch field.Type {
			case queryType:
				println("query")
				w.queryType = inParam
				return
			case bodyType:
				println("body")
				w.bodyType = inParam
				return
			}
		}
	}
}



func (w *HandlerWrapper) rawHandle(ctx *gin.Context) {

}
