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

type HandlerWrapper struct {
	originalHandler interface{}
	handlersChain   []*HandlerCaller
	pathParams      []reflect.Value
	valueNum        int
	valueTypes      map[reflect.Type]int
	queryType       *reflect.Type
	bodyType        *reflect.Type
	docs            interface{}
	method          string
	path            string
	middlewares     []Middleware
	params          pathParameters
	errorIndex      int
}

func (w *HandlerWrapper) init() {
	w.docs = gin.H{"elo": "ole"}

	for _, middleware := range w.middlewares {
		if middleware.Before != nil {
			w.chainHandler(middleware.Before)
		}
	}

	w.chainHandler(w.originalHandler)

	for i := len(w.middlewares) - 1; i >= 0; i-- {
		middleware := w.middlewares[i]
		if middleware.After != nil {
			w.chainHandler(middleware.After)
		}
	}
}

func (w *HandlerWrapper) chainHandler(handler interface{}) {
	caller := NewHandlerCaller(reflect.ValueOf(handler))

	ht := reflect.TypeOf(handler)
	if ht.Kind() != reflect.Func {
		panic("handler is not a function")
	}

	w.routeInParams(ht, caller)
	w.routeOutParams(ht, caller)

	w.handlersChain = append(w.handlersChain, caller)
}

func (w *HandlerWrapper) routeInParams(ht reflect.Type, caller *HandlerCaller) {
	paramIndex := 0
	for i := 0; i < ht.NumIn(); i++ {
		inParam := ht.In(i)

		if w.isPathParam(inParam) {
			w.addPathParamBuilder(caller, inParam, paramIndex)
			paramIndex++
		}

		if index, exists := w.valueTypes[inParam]; exists {
			caller.addBuilder(cachedValue(index))
			continue
		}

		w.valueTypes[inParam] = w.valueNum
		w.valueNum++

		switch {
		case inParam.Implements(bodyType):
			w.setBodyType(inParam)
			w.addGenericBuilder(caller, inParam, binding.JSON)
		case inParam.Implements(queryType):
			w.setQueryType(inParam)
			w.addGenericBuilder(caller, inParam, binding.Query)
		case inParam.Implements(headersType):
			caller.addBuilder(headersBuilder(inParam.Kind() == reflect.Ptr))
			return
		default:
			switch w.method {
			case http.MethodGet, http.MethodDelete, http.MethodHead, http.MethodOptions:
				w.setQueryType(inParam)
				w.addGenericBuilder(caller, inParam, binding.Query)
			case http.MethodPost, http.MethodPatch, http.MethodPut:
				w.setBodyType(inParam)
				w.addGenericBuilder(caller, inParam, binding.JSON)
			default:
				panic("unknown parameter purpose; neither body nor query")
			}
		}
	}
}

func (w *HandlerWrapper) routeOutParams(ht reflect.Type, caller *HandlerCaller) {
	paramIndex := 0
	for i := 0; i < ht.NumOut(); i++ {
		outParam := ht.Out(i)

		optional := false
		if outParam.Kind() == reflect.Ptr {
			optional = true
			outParam = outParam.Elem()
		}

		switch outParam.Kind() {
		case reflect.Int, reflect.String:
			caller.addBuilder(paramBuilder(outParam.Kind(), w.params.index(paramIndex), optional))
			paramIndex++
		case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
			w.addGenericBuilder(caller, outParam, optional)
		case reflect.Ptr:
			panic("can not use double pointer as an argument: " + ht.In(i).String())
		default:
			panic("unknown parameter kind")
		}
	}
}

func (w *HandlerWrapper) rawHandle(ctx *gin.Context) {
	context := &callContext{
		rawContext: ctx,
		values:     make([]*reflect.Value, w.valueNum),
	}

	for _, caller := range w.handlersChain {
		err := caller.call(context)
		if err != nil {
			break
		}
	}

	response := context.response
	if context.error != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, context.error)
		return
	}

	ctx.JSON(200, response.Interface())
}

func (w *HandlerWrapper) setBodyType(argType reflect.Type) {
	if w.bodyType != nil {
		panic(fmt.Sprintf("ambiguous body parameter: %s and %s", (*w.bodyType).Name(), argType.Name()))
	}
	w.bodyType = &argType
}

func (w *HandlerWrapper) setQueryType(argType reflect.Type) argBuilder {
	if w.queryType != nil {
		panic(fmt.Sprintf("ambiguous query parameter: %s and %s", (*w.queryType).Name(), argType.Name()))
	}
	w.queryType = &argType
	return genericBuilder(argType, binding.Query)
}

func (w *HandlerWrapper) isPathParam(argType reflect.Type) bool {
	if argType.Kind() == reflect.Ptr {
		argType = argType.Elem()
	}
	return argType == intType || argType == stringType
}

func (w *HandlerWrapper) addPathParamBuilder(caller *HandlerCaller, argType reflect.Type, paramIndex int) {
	optional := false
	if argType.Kind() == reflect.Ptr {
		optional = true
		argType = argType.Elem()
	}
	caller.addBuilder(paramBuilder(argType.Kind(), w.params.index(paramIndex), optional))
}

func (w *HandlerWrapper) addGenericBuilder(caller *HandlerCaller, argType reflect.Type, bindType binding.Binding) {
	caller.addBuilder(cached(genericBuilder(argType, bindType), w.valueNum))
}
