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
		valuesTypes:     map[reflect.Type]int{},
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
	valuesNum       int
	valuesTypes     map[reflect.Type]int
	queryType       *reflect.Type
	bodyType        *reflect.Type
	responseType    *reflect.Type
	docs            interface{}
	method          string
	path            string
	middlewares     []Middleware
	params          pathParameters
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

	w.inspectInParams(ht, caller)
	w.inspectOutParams(ht, caller)

	w.handlersChain = append(w.handlersChain, caller)
}

func (w *HandlerWrapper) inspectInParams(handlerType reflect.Type, caller *HandlerCaller) {
	paramIndex := 0
	for i := 0; i < handlerType.NumIn(); i++ {
		arg := handlerType.In(i)

		if w.isPathParam(arg) {
			w.addPathParamBuilder(caller, arg, paramIndex)
			paramIndex++
			continue
		}

		if index, exists := w.valuesTypes[arg]; exists {
			caller.addBuilder(cachedValue(index))
			continue
		}
		
		switch {
		case arg.Implements(bodyInterfaceType):
			w.setBodyType(arg)
			w.addGenericBuilder(caller, arg, binding.JSON)
		case arg.Implements(queryInterfaceType):
			w.setQueryType(arg)
			w.addGenericBuilder(caller, arg, binding.Query)
		case arg.Implements(headersInterfaceType):
			w.addGenericBuilder(caller, arg, binding.Header)
		default:
			switch w.method {
			case http.MethodGet, http.MethodDelete, http.MethodHead, http.MethodOptions:
				w.setQueryType(arg)
				w.addGenericBuilder(caller, arg, binding.Query)
			case http.MethodPost, http.MethodPatch, http.MethodPut:
				w.setBodyType(arg)
				w.addGenericBuilder(caller, arg, binding.JSON)
			default:
				panic("unknown parameter purpose; neither body nor query")
			}
		}

		w.valuesTypes[arg] = w.valuesNum
		w.valuesNum++
	}
}

func (w *HandlerWrapper) inspectOutParams(handlerType reflect.Type, caller *HandlerCaller) {
	for i := 0; i < handlerType.NumOut(); i++ {
		arg := handlerType.Out(i)

		if index, exists := w.valuesTypes[arg]; exists {
			caller.addSetter(valueSetter(index))
			continue
		}


		switch {
		case arg.Implements(responseInterfaceType):
			w.setResponseType(arg)
			caller.addSetter(valueSetter(w.valuesNum))
		case arg.Implements(errorInterfaceType):
			caller.addSetter(errorSetter(arg.Kind() == reflect.Ptr))
		case arg == headersType || (arg.Kind() == reflect.Ptr && arg.Elem() == headersType):
			caller.addSetter(headersSetter(arg.Kind() == reflect.Ptr))
			// Do not cache response headers
			continue
		case 
		default:
			caller.addSetter(valueSetter(w.valuesNum))
		}

		w.valuesTypes[arg] = w.valuesNum
		w.valuesNum++
	}
}

func (w *HandlerWrapper) rawHandle(ctx *gin.Context) {
	context := &callContext{
		rawContext: ctx,
		values:     make([]*reflect.Value, w.valuesNum),
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
		panic(fmt.Sprintf("ambiguous body type: %s and %s", *w.bodyType, argType))
	}
	w.bodyType = &argType
}

func (w *HandlerWrapper) setQueryType(argType reflect.Type) {
	if w.queryType != nil {
		panic(fmt.Sprintf("ambiguous query type: %s and %s", *w.queryType, argType))
	}
	w.queryType = &argType
}

func (w *HandlerWrapper) setResponseType(argType reflect.Type) {
	if w.responseType != nil {
		panic(fmt.Sprintf("ambiguous response type: %s and %s", *w.responseType, argType))
	}
	w.responseType = &argType
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
	caller.addBuilder(cached(genericBuilder(argType, bindType), w.valuesNum))
}
