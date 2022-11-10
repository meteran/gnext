package gnext

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/meteran/gnext/docs"
	"net/http"
	"reflect"
	"regexp"
)

var paramRegExp = regexp.MustCompile(":[a-zA-Z0-9]+/")

type handlerType string

const (
	htBeforeMiddleware handlerType = "before"
	htTargetHandler    handlerType = "target"
	htAfterMiddleware  handlerType = "after"
	htErrorHandler     handlerType = "error"
)

func WrapHandler(method string, path string, middlewares middlewares, documentation *docs.Docs, handler, errorHandler interface{}, doc ...*docs.Endpoint) *HandlerWrapper {
	wrapper := &HandlerWrapper{
		method:          method,
		path:            path,
		middlewares:     middlewares,
		originalHandler: handler,
		errorHandler:    errorHandler,
		docs:            documentation,
		params:          newParameters(path),
		valuesTypes:     map[reflect.Type]int{},
		responseIndex:   -1,
		defaultStatus:   200,
	}

	if len(doc) == 0 {
		wrapper.doc = &docs.Endpoint{}
	} else {
		wrapper.doc = doc[0]
	}

	wrapper.init()
	if wrapper.documentedRouter() {
		wrapper.fillDocumentation()
	}
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
	originalHandler    interface{}
	handlersChain      []*HandlerCaller
	handlerFallbacks   []int
	pathParams         []reflect.Value
	errorHandler       interface{}
	errorHandlerCaller *ErrorHandlerCaller
	valuesNum          int
	valuesTypes        map[reflect.Type]int
	queryType          reflect.Type
	bodyType           reflect.Type
	headerTypes        []reflect.Type
	responseType       reflect.Type
	docs               *docs.Docs
	doc                *docs.Endpoint
	method             string
	path               string
	middlewares        middlewares
	params             pathParameters
	responseIndex      int
	defaultStatus      Status
}

func (w *HandlerWrapper) documentedRouter() bool {
	return w.docs != nil
}

func (w *HandlerWrapper) init() {
	middlewaresCount := w.middlewares.count()
	lastAfterMiddleware := -1
	for _, middleware := range w.middlewares {
		if middleware.After != nil {
			if lastAfterMiddleware == -1 {
				lastAfterMiddleware = middlewaresCount
			} else {
				lastAfterMiddleware--
			}
		}
		if middleware.Before != nil {
			w.chainHandler(middleware.Before, htBeforeMiddleware)
			w.handlerFallbacks = append(w.handlerFallbacks, lastAfterMiddleware)
		}
	}

	w.chainHandler(w.originalHandler, htTargetHandler)
	w.handlerFallbacks = append(w.handlerFallbacks, lastAfterMiddleware)

	w.wrapErrorHandler()

	for i := len(w.middlewares) - 1; i >= 0; i-- {
		middleware := w.middlewares[i]
		if middleware.After != nil {
			lastAfterMiddleware++
			if lastAfterMiddleware > middlewaresCount {
				lastAfterMiddleware = -1
			}
			w.chainHandler(middleware.After, htAfterMiddleware)
			w.handlerFallbacks = append(w.handlerFallbacks, lastAfterMiddleware)
		}
	}
}

func (w *HandlerWrapper) chainHandler(handler interface{}, hType handlerType) {
	caller := NewHandlerCaller(reflect.ValueOf(handler))

	ht := reflect.TypeOf(handler)
	if ht.Kind() != reflect.Func {
		panic(fmt.Sprintf("'%s' is not a function", ht))
	}

	w.inspectInParams(ht, caller, hType)
	w.inspectOutParams(ht, caller, hType)

	w.handlersChain = append(w.handlersChain, caller)
}

func (w *HandlerWrapper) inspectInParams(handlerType reflect.Type, caller *HandlerCaller, hType handlerType) {
	paramIndex := 0
	for i := 0; i < handlerType.NumIn(); i++ {
		arg := handlerType.In(i)

		switch {
		case w.isPathParam(arg):
			w.addPathParamBuilder(caller, arg, paramIndex)
			if w.documentedRouter() {
				w.doc.AddPathParam(w.params.index(paramIndex), arg)
			}
			paramIndex++

			continue
		case typesEqual(statusType, arg):
			caller.addBuilder(statusBuilder(isPtr(arg)))
			continue
		}

		if index, exists := w.valuesTypes[arg]; exists {
			builder := cachedValue(index)

			// if an error occurred, it might be that the needed input value is unset
			// in such case we need to initiate it with a zero value to
			if hType == htAfterMiddleware {
				builder = optionallyCachedValue(index, arg)
			}
			caller.addBuilder(builder)
			continue
		}

		switch {
		case arg == rawContextType:
			caller.addBuilder(cached(rawContextBuilder, w.valuesNum))
		case arg.Implements(bodyInterfaceType):
			w.setBodyType(arg)
			w.addGenericBuilder(caller, arg, binding.JSON)
		case arg.Implements(queryInterfaceType):
			w.setQueryType(arg)
			w.addGenericBuilder(caller, arg, binding.Query)
		case arg.Implements(headersInterfaceType):
			w.appendHeadersType(arg)
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
				panic("unknown input parameter purpose or type; allowed values are: request body, query and path params, headers or one of the types returned from previous middlewares")
			}
		}

		w.valuesTypes[arg] = w.valuesNum
		w.valuesNum++
	}
}

type producingCaller interface {
	addSetter(setter argSetter)
}

func (w *HandlerWrapper) inspectOutParams(handlerType reflect.Type, caller producingCaller, hType handlerType) {
	for i := 0; i < handlerType.NumOut(); i++ {
		arg := handlerType.Out(i)

		// response parameters that shouldn't be in cache
		switch {
		case typesEqual(headersType, arg):
			caller.addSetter(headersSetter(isPtr(arg)))
			continue
		case typesEqual(statusType, arg):
			caller.addSetter(statusSetter(isPtr(arg)))
			continue
		case arg.Implements(errorInterfaceType):
			caller.addSetter(errorSetter)
			continue
		}

		if index, exists := w.valuesTypes[arg]; exists {
			caller.addSetter(valueSetter(index))
			continue
		}

		switch {
		// if this is a target handler
		// we consider any unknown returned object as a response
		// just for developer convenience
		case arg.Implements(responseInterfaceType) || hType == htTargetHandler:
			w.setResponseType(arg)
			caller.addSetter(valueSetter(w.valuesNum))
		default:
			caller.addSetter(valueSetter(w.valuesNum))
		}

		w.valuesTypes[arg] = w.valuesNum
		w.valuesNum++
	}
}

func (w *HandlerWrapper) appendHeadersType(argType reflect.Type) {
	if argType.Kind() != reflect.Map {
		w.headerTypes = append(w.headerTypes, argType)
	}
}

func (w *HandlerWrapper) setBodyType(argType reflect.Type) {
	if w.bodyType != nil {
		panic(fmt.Sprintf("ambiguous body type: %s and %s", w.bodyType, argType))
	}
	w.bodyType = argType
}

func (w *HandlerWrapper) setQueryType(argType reflect.Type) {
	if w.queryType != nil {
		panic(fmt.Sprintf("ambiguous query type: %s and %s", w.queryType, argType))
	}
	w.queryType = argType
}

func (w *HandlerWrapper) setResponseType(argType reflect.Type) {
	if w.responseType != nil {
		panic(fmt.Sprintf("ambiguous response type: %s and %s", w.responseType, argType))
	}
	w.responseType = argType
	w.responseIndex = w.valuesNum
}

func (w *HandlerWrapper) isPathParam(argType reflect.Type) bool {
	if isPtr(argType) {
		argType = argType.Elem()
	}
	return argType == intType || argType == stringType
}

func (w *HandlerWrapper) addPathParamBuilder(caller *HandlerCaller, argType reflect.Type, paramIndex int) {
	optional := false
	if isPtr(argType) {
		optional = true
		argType = argType.Elem()
	}
	caller.addBuilder(paramBuilder(argType.Kind(), w.params.index(paramIndex), optional))
}

func (w *HandlerWrapper) addGenericBuilder(caller *HandlerCaller, argType reflect.Type, bindType binding.Binding) {
	caller.addBuilder(cached(genericBuilder(argType, bindType), w.valuesNum))
}

func (w *HandlerWrapper) fillDocumentation() {
	w.doc.SetTagsFromPath(w.path)

	if w.bodyType != nil {
		w.doc.SetBodyType(w.bodyType)
	}

	if w.responseType != nil {
		w.doc.SetResponses(w.responseType, w.errorHandlerCaller.responseType)
		w.defaultStatus = Status(docs.DefaultStatus(w.responseType))
	}

	if w.queryType != nil {
		w.doc.SetQueryType(w.queryType)
	}

	for _, headerType := range w.headerTypes {
		w.doc.AddHeadersType(headerType)
	}

	w.docs.SetPath(w.path, w.method, w.doc)
}

func (w *HandlerWrapper) requestHandler(rawContext *gin.Context) {
	context := &callContext{
		rawContext: rawContext,
		values:     make([]*reflect.Value, w.valuesNum),
		status:     w.defaultStatus,
	}

	for i := 0; i < len(w.handlersChain); {
		w.handlersChain[i].call(context)
		if context.error != nil {
			w.errorHandlerCaller.call(context)
			i = w.handlerFallbacks[i]
			if i < 0 {
				break
			}
		} else {
			i++
		}
	}

	response := context.values[w.responseIndex]
	if response == nil {
		rawContext.AbortWithStatus(int(context.status))
		return
	}
	rawContext.JSON(int(context.status), response.Interface())
}

func (w *HandlerWrapper) withoutResponseRequestHandler(rawContext *gin.Context) {
	context := &callContext{
		rawContext: rawContext,
		values:     make([]*reflect.Value, w.valuesNum),
		status:     w.defaultStatus,
	}

	for i := 0; i < len(w.handlersChain); {
		w.handlersChain[i].call(context)
		if context.error != nil {
			w.errorHandlerCaller.call(context)
			i = w.handlerFallbacks[i]
			if i < 0 {
				break
			}
		} else {
			i++
		}
	}

	rawContext.AbortWithStatus(int(context.status))
}

func (w *HandlerWrapper) getHandler() gin.HandlerFunc {
	if w.responseIndex < 0 {
		return w.withoutResponseRequestHandler
	}
	return w.requestHandler
}

func (w *HandlerWrapper) wrapErrorHandler() {
	w.errorHandlerCaller = &ErrorHandlerCaller{
		originalHandler: w.errorHandler,
	}

	w.errorHandlerCaller.init()
}
