package gnext

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"os"
	"reflect"
)

var resetColor = "\033[0m"
var errLog = log.New(os.Stderr, "\n\n\x1b[31m", log.LstdFlags)

// errorHandlers is a mapping from an error type to that error handler
type errorHandlers map[reflect.Type]reflect.Value

func (h errorHandlers) setup(handler interface{}) {
	ht := reflect.TypeOf(handler)
	validateErrorHandler(ht)

	h[ht.In(0)] = reflect.ValueOf(handler)
}

func (h errorHandlers) copy() errorHandlers {
	newHandlers := make(errorHandlers, len(h))
	for typ, value := range h {
		newHandlers[typ] = value
	}
	return newHandlers
}

func newErrorHandlers() errorHandlers {
	handlers := errorHandlers{}
	handlers.setup(DefaultErrorHandler)
	return handlers
}

func DefaultErrorHandler(err error) (status Status, response *DefaultErrorResponse) {
	status = 500
	response = &DefaultErrorResponse{
		Success: false,
		Message: "internal server error",
	}

	switch e := err.(type) {
	case *json.SyntaxError:
		status = http.StatusBadRequest
		response.Message = "malformed json"
		response.Details = []string{fmt.Sprintf("at position %d: %s", e.Offset, e.Error())}
	case *json.UnmarshalTypeError:
		status = http.StatusBadRequest
		response.Message = "invalid payload"
		response.Details = []string{fmt.Sprintf("at position %d: invalid type of field '%s': was '%s', should be '%s'", e.Offset, e.Field, e.Value, e.Type.Name())}
	case validator.ValidationErrors:
		status = http.StatusBadRequest
		response.Message = "validation error"
		for _, validationError := range e {
			response.Details = append(response.Details, fmt.Sprintf("field validation for '%s' failed on the '%s' tag with value '%s'",
				validationError.Field(), validationError.ActualTag(), validationError.Param()))
		}
	case *NotFound:
		status = http.StatusNotFound
		response.Message = err.Error()
	case *HandlerPanicked:
		errLog.Printf("panic recovered: %v\n%s%s", e.Value, e.StackTrace, resetColor)
	default:
		// TODO if debug mode then return `err.Error()` in message
		errLog.Printf("unhandled error: %v%s", err, resetColor)
	}
	return
}

func validateErrorHandler(ht reflect.Type) {
	if ht.Kind() != reflect.Func {
		panic(fmt.Sprintf("error handler '%s' is not a function", ht))
	}

	if ht.NumIn() != 1 {
		panic(fmt.Sprintf("error handler '%s' must accept exactly one argument implementing 'error', was '%d' arguments", ht, ht.NumIn()))
	}

	input := ht.In(0)
	if !input.Implements(errorInterfaceType) {
		panic(fmt.Sprintf("error handler '%s' must accept argument implementing 'error`, got type '%s'", ht, input))
	}
}

func newErrorHandlerCaller(handler reflect.Value) *errorHandlerCaller {
	return &errorHandlerCaller{
		handler:       handler,
		defaultStatus: 500,
	}
}

type errorHandlerCaller struct {
	handler       reflect.Value
	argSetters    []argSetter
	defaultStatus Status
}

func (c *errorHandlerCaller) addSetter(setter argSetter) {
	c.argSetters = append(c.argSetters, setter)
}

func (c *errorHandlerCaller) call(ctx *callContext) {
	results := c.handler.Call([]reflect.Value{ctx.error.Elem()})
	ctx.status = c.defaultStatus

	for i, setter := range c.argSetters {
		setter(&results[i], ctx)
	}
}
