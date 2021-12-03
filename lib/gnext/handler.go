package gnext

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"reflect"
	"regexp"
)

var paramRegExp = regexp.MustCompile(":[a-zA-Z0-9]+/")

func WrapHandler(method string, path string, handler interface{}) *HandlerWrapper {
	wrapper := &HandlerWrapper{
		method:          method,
		path:            path,
		originalHandler: handler,
	}
	wrapper.init()
	return wrapper
}

func newParameters(path string) parameters {
	return parameters{
		paramNames:     paramRegExp.FindAll([]byte(path), -1),
		nextParamIndex: 0,
	}
}

type parameters struct {
	paramNames     [][]byte
	nextParamIndex int
}

func (p *parameters) next() string {
	if p.nextParamIndex >= len(p.paramNames) {
		panic(fmt.Sprintf("too many path parameters in handler; currently in path: %d", len(p.paramNames)))
	}
	p.nextParamIndex++
	paramName := string(p.paramNames[p.nextParamIndex-1])
	return paramName[1:len(paramName)-1]
}

type HandlerWrapper struct {
	originalHandler interface{}
	handlerValue    reflect.Value
	paramBuilders   []builder
	pathParams      []reflect.Value
	queryType       *reflect.Type
	bodyType        *reflect.Type
	docs            interface{}
	method          string
	path            string
}

func (w *HandlerWrapper) init() {
	w.docs = gin.H{"elo": "ole"}

	params := newParameters(w.path)

	w.handlerValue = reflect.ValueOf(w.originalHandler)
	ht := reflect.TypeOf(w.originalHandler)
	if ht.Kind() != reflect.Func {
		panic("handler is not a function")
	}

	for i := 0; i < ht.NumIn(); i++ {
		inParam := ht.In(i)

		switch inParam.Kind() {
		case reflect.Int:
			w.addBuilder(intParamBuilder(params.next()))
		case reflect.String:
			w.addBuilder(stringParamBuilder(params.next()))
		case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
			w.appendGeneric(inParam, false)
		case reflect.Ptr:
			innerType := inParam.Elem()
			switch innerType.Kind() {
			case reflect.Int:
				w.addBuilder(intOptionalParamBuilder(params.next()))
			case reflect.String:
				w.addBuilder(stringOptionalParamBuilder(params.next()))
			case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
				w.appendGeneric(innerType, true)
			default:
				panic("unknown optional parameter")
			}
		default:
			panic("unknown parameter")
		}
	}
}

func (w *HandlerWrapper) addBuilder(b builder) {
	w.paramBuilders = append(w.paramBuilders, b)
}

func (w *HandlerWrapper) appendGeneric(inParam reflect.Type, optional bool) {
	var paramType string
	switch {
	case inParam.Implements(bodyType):
		paramType = BodyParam
	case inParam.Implements(queryType):
		paramType = QueryParam
	case inParam == headersType:
		w.addBuilder(headerBuilder(optional))
		return
	default:
		switch w.method {
		case http.MethodGet, http.MethodDelete, http.MethodHead, http.MethodOptions:
			paramType = QueryParam
		case http.MethodPost, http.MethodPatch, http.MethodPut:
			paramType = BodyParam
		default:
			panic("unknown parameter purpose; neither body nor query")
		}
	}

	var format binding.Binding
	switch paramType {
	case BodyParam:
		if w.bodyType != nil {
			panic(fmt.Sprintf("ambiguous body parameter: %s and %s", (*w.bodyType).Name(), inParam.Name()))
		}
		w.bodyType = &inParam
		format = binding.JSON
	case QueryParam:
		if w.queryType != nil {
			panic(fmt.Sprintf("ambiguous query parameter: %s and %s", (*w.queryType).Name(), inParam.Name()))
		}
		w.queryType = &inParam
		format = binding.Query
	}

	w.addBuilder(genericBuilder(inParam, format, optional))
}

func (w *HandlerWrapper) rawHandle(ctx *gin.Context) {
	values := make([]reflect.Value, len(w.paramBuilders))
	for i, builder := range w.paramBuilders {
		value, err := builder(ctx)
		if err != nil {
			fmt.Printf("error parsing request: %v\n", err)
			panic("unhandled errors for now")
		}
		values[i] = value
	}

	result := w.handlerValue.Call(values)
	response := result[0]
	err := result[1]
	if !err.IsNil() {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err.Interface())
		return
	}
	ctx.JSON(200, response.Interface())
}
