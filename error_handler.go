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
		panic(fmt.Sprintf("error handler '%s' must accept exactly one argument of type 'error', was '%d' arguments", ht, ht.NumIn()))
	}

	input := ht.In(0)
	if !input.Implements(errorInterfaceType) {
		panic(fmt.Sprintf("error handler '%s' must accept argument of type 'error' or implements 'error` instead of type '%s'", ht, input))
	}

	//if ht.NumOut() != 2 {
	//	panic(fmt.Sprintf("error handler '%s' must return exactly two arguments(gnext.Status and a response object), was '%d' arguments", ht, ht.NumOut()))
	//}
}

func newErrorHandlerCaller(handler interface{}) *ErrorHandlerCaller {
	return &ErrorHandlerCaller{
		originalHandler: handler,
		handler:         reflect.ValueOf(handler),
	}
}

type ErrorHandlerCaller struct {
	originalHandler interface{}
	handler         reflect.Value
	argSetters      []argSetter
}

func (c *ErrorHandlerCaller) addSetter(setter argSetter) {
	c.argSetters = append(c.argSetters, setter)
}

func (c *ErrorHandlerCaller) call(ctx *callContext) {
	results := c.handler.Call([]reflect.Value{*ctx.error})

	//status := results[c.statusIndex].Convert(intType).Interface().(int)
	//response := results[c.responseIndex].Interface()
	//if status == 0 {
	//	status = 500
	//	response = &DefaultErrorResponse{
	//		Message: "internal server error",
	//		Success: false,
	//	}
	//}

	for i, setter := range c.argSetters {
		setter(&results[i], ctx)
	}
}
