package gnext

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gnext.io/gnext/docs"
	"net/http"
	"reflect"
	"regexp"
)

var paramRegExp = regexp.MustCompile(":[a-zA-Z0-9]+/")

func WrapHandler(method string, path string, middlewares []Middleware, docs *docs.Docs, handler interface{}) *HandlerWrapper {
	wrapper := &HandlerWrapper{
		method:          method,
		path:            path,
		middlewares:     middlewares,
		originalHandler: handler,
		docs:            docs,
		params:          newParameters(path),
		valuesTypes:     map[reflect.Type]int{},
		defaultStatus:   200,
	}
	wrapper.init()
	wrapper.fillDocumentation()
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
	queryType       reflect.Type
	bodyType        reflect.Type
	responseType    reflect.Type
	docs            *docs.Docs
	method          string
	path            string
	middlewares     []Middleware
	params          pathParameters
	responseIndex   int
	defaultStatus   Status
}

func (w *HandlerWrapper) init() {

	for _, middleware := range w.middlewares {
		if middleware.Before != nil {
			w.chainHandler(middleware.Before, false)
		}
	}

	w.chainHandler(w.originalHandler, true)

	for i := len(w.middlewares) - 1; i >= 0; i-- {
		middleware := w.middlewares[i]
		if middleware.After != nil {
			w.chainHandler(middleware.After, false)
		}
	}
}

func (w *HandlerWrapper) fillDocumentation() {
	docBuilder := docs.NewBuilder(w.docs)
	var operation openapi3.Operation

	operation.Tags = docBuilder.Helper.GetTagsFromPath(w.path)

	if w.bodyType != nil {
		bodyModel := docBuilder.Helper.ConvertTypeToInterface(w.bodyType.Elem())
		operation.RequestBody = docBuilder.Helper.ConvertModelToRequestBody(bodyModel, "")
	}

	if w.responseType != nil {
		responseModel := docBuilder.Helper.ConvertTypeToInterface(w.responseType.Elem())
		operation.Responses = docBuilder.Helper.CreateResponses(responseModel, nil)
	}

	docBuilder.Helper.AddParametersToOperation(docBuilder.Helper.ParsePathParams(w.path), &operation)

	if w.queryType != nil {
		queryModel := docBuilder.Helper.ConvertTypeToInterface(w.queryType.Elem())
		docBuilder.Helper.AddParametersToOperation(docBuilder.Helper.ParseQueryParams(queryModel), &operation)
	}

	docBuilder.SetOperationOnPath(w.path, w.method, operation)
}

func (w *HandlerWrapper) chainHandler(handler interface{}, final bool) {
	caller := NewHandlerCaller(reflect.ValueOf(handler))

	ht := reflect.TypeOf(handler)
	if ht.Kind() != reflect.Func {
		panic("handler is not a function")
	}

	w.inspectInParams(ht, caller)
	w.inspectOutParams(ht, caller, final)

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
		case arg == rawContextType:
			caller.addBuilder(cached(rawContextBuilder, w.valuesNum))
		case typesEqual(statusType, arg):
			caller.addBuilder(statusBuilder(arg.Kind() == reflect.Ptr))
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

func (w *HandlerWrapper) inspectOutParams(handlerType reflect.Type, caller *HandlerCaller, final bool) {
	for i := 0; i < handlerType.NumOut(); i++ {
		arg := handlerType.Out(i)

		// response parameters that shouldn't be in cache
		switch {
		case typesEqual(headersType, arg):
			caller.addSetter(headersSetter(arg.Kind() == reflect.Ptr))
			continue
		case typesEqual(statusType, arg):
			caller.addSetter(statusSetter(arg.Kind() == reflect.Ptr))
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
		// final means, that it's the original handler
		// in such case we consider any unknown returned object as a response
		// just for developer convenience
		case arg.Implements(responseInterfaceType) || final:
			w.setResponseType(arg)
			caller.addSetter(valueSetter(w.valuesNum))
		default:
			caller.addSetter(valueSetter(w.valuesNum))
		}

		w.valuesTypes[arg] = w.valuesNum
		w.valuesNum++
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

func (w *HandlerWrapper) rawHandle(rawContext *gin.Context) {
	context := &callContext{
		rawContext: rawContext,
		values:     make([]*reflect.Value, w.valuesNum),
		status:     w.defaultStatus,
	}

	for _, caller := range w.handlersChain {
		err := caller.call(context)
		if err != nil {
			break
		}
	}

	response := context.values[w.responseIndex]
	if context.error != nil {
		rawContext.AbortWithStatusJSON(http.StatusInternalServerError, context.error)
		return
	}

	if response == nil {
		rawContext.AbortWithStatus(int(context.status))
		return
	}
	rawContext.JSON(int(context.status), response.Interface())
}
