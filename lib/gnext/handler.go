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

func WrapHandler(method string, path string, middlewares []Middleware, handler interface{}) *HandlerWrapper {
	wrapper := &HandlerWrapper{
		method:          method,
		path:            path,
		middlewares:     middlewares,
		originalHandler: handler,
		params:          newParameters(path),
	}
	wrapper.init()
	return wrapper
}

func newParameters(path string) pathParameters {
	return pathParameters{
		paramNames: paramRegExp.FindAll([]byte(path), -1),
	}
}

type pathParameters struct {
	paramNames [][]byte
}

func (p *pathParameters) index(index int) string {
	if index >= len(p.paramNames) {
		panic(fmt.Sprintf("path parameter index out of range: %d", len(p.paramNames)))
	}
	paramName := string(p.paramNames[index])
	return paramName[1 : len(paramName)-1]
}

type argSetter func([]reflect.Value, []*reflect.Value)

type HandlerCaller struct {
	argBuilders []builder
	argSetters  []argSetter
	receiver    reflect.Value
}

func (c *HandlerCaller) addSetter(outputIndex, contextIndex int) {
	setter := func(output []reflect.Value, context []*reflect.Value) {
		context[contextIndex] = &output[outputIndex]
	}
	c.argSetters = append(c.argSetters, setter)
}

func (c *HandlerCaller) addBuilder(b builder) {
	c.argBuilders = append(c.argBuilders, b)
}

func (c *HandlerCaller) Call(contextValues []*reflect.Value, ctx *gin.Context) *Error {
	values := make([]reflect.Value, len(c.argBuilders))
	for i, builder := range c.argBuilders {
		value, err := builder(ctx, contextValues)
		if err != nil {
			return &Error{}
		}
		values[i] = value
	}

	result := c.receiver.Call(values)
	for _, setter := range c.argSetters {
		setter(result, contextValues)
	}
	return nil
}

type HandlerWrapper struct {
	originalHandler interface{}
	handlersChain   []*HandlerCaller
	pathParams      []reflect.Value
	queryType       *reflect.Type
	bodyType        *reflect.Type
	docs            interface{}
	method          string
	path            string
	middlewares     []Middleware
	params          pathParameters
	responseIndex   int
	errorIndex      int
	valuesNum       int
}

func (w *HandlerWrapper) init() {
	w.docs = gin.H{"elo": "ole"}

	for _, middleware := range w.middlewares {
		if middleware.Before != nil {
			w.chainHandler(middleware.Before)
		}
	}

}

func (w *HandlerWrapper) chainHandler(handler interface{}) {
	caller := &HandlerCaller{
		receiver: reflect.ValueOf(handler),
	}

	ht := reflect.TypeOf(handler)
	if ht.Kind() != reflect.Func {
		panic("handler is not a function")
	}

	paramIndex := 0
	for i := 0; i < ht.NumIn(); i++ {
		inParam := ht.In(i)

		optional := false
		if inParam.Kind() == reflect.Ptr {
			optional = true
			inParam = inParam.Elem()
		}

		switch inParam.Kind() {
		case reflect.Int, reflect.String:
			w.addBuilder(paramBuilder(inParam.Kind(), w.params.index(paramIndex), optional))
			paramIndex++
		case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
			w.appendGeneric(inParam, optional)
		case reflect.Ptr:
			panic("can not use double pointer as parameter: " + ht.In(i).String())
		default:
			panic("unknown parameter")
		}
	}
	w.handlersChain = append(w.handlersChain, caller)
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
	contextValues := make([]*reflect.Value, w.valuesNum)

	for _, caller := range w.handlersChain {
		err := caller.Call(contextValues, ctx)
		if err != nil {
			fmt.Printf("error parsing request: %v\n", err)
			panic("unhandled errors for now")
		}
	}

	response := contextValues[w.responseIndex]
	err := contextValues[w.errorIndex]
	if !err.IsNil() {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err.Interface())
		return
	}
	ctx.JSON(200, response.Interface())
}
