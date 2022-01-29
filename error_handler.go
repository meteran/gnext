package gnext

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"reflect"
)

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
		(&spew.ConfigState{
			Indent:                  " ",
			DisablePointerAddresses: true,
			DisableCapacities:       true,
			SortKeys:                true,
			DisableMethods:          true,
			MaxDepth:                2,
		}).Dump(e[0])
		status = http.StatusBadRequest
		response.Message = "validation error"
		for _, validationError := range e {
			response.Details = append(response.Details, validationError.Error())
		}
	case *NotFound:
		status = http.StatusNotFound
		response.Message = err.Error()
	default:
		// TODO if debug mode then return `err.Error()` in message
		log.Printf("unhandled error: %v", err)
	}
	return
}

func WrapErrorHandler(handler interface{}) *ErrorHandlerCaller {
	caller := &ErrorHandlerCaller{
		originalHandler: handler,
	}

	caller.init()
	return caller
}

type ErrorHandlerCaller struct {
	originalHandler interface{}
	handler         reflect.Value
	responseType    reflect.Type
	responseIndex   int
	statusIndex     int
}

func (c *ErrorHandlerCaller) init() {
	c.handler = reflect.ValueOf(c.originalHandler)

	ht := reflect.TypeOf(c.originalHandler)
	c.validate(ht)
	c.recognizeOutParams(ht)
}

func (c *ErrorHandlerCaller) validate(ht reflect.Type) {
	if ht.Kind() != reflect.Func {
		panic(fmt.Sprintf("error handler '%s' is not a function", ht))
	}

	if ht.NumIn() != 1 {
		panic(fmt.Sprintf("error handler '%s' must accept exactly one argument of type 'error', was '%d' arguments", ht, ht.NumIn()))
	}

	if ht.In(0) != errorInterfaceType {
		panic(fmt.Sprintf("error handler '%s' must accept argument of type 'error' instead of type '%s'", ht, ht.In(0)))
	}

	if ht.NumOut() != 2 {
		panic(fmt.Sprintf("error handler '%s' must return exactly two arguments, was '%d' arguments", ht, ht.NumOut()))
	}
}

func (c *ErrorHandlerCaller) recognizeOutParams(ht reflect.Type) {
	errorMsg := fmt.Sprintf("error handler '%s' must return status of type 'int' and some other type as response payload; if you need an integer as payload, use 'gnext.Status' for status", ht)

	switch {
	// NOTE: order of cases below is meaningful, do not touch it without a good reason
	case ht.Out(0) == ht.Out(1):
		panic(errorMsg)
	case ht.Out(0) == statusType:
		c.statusIndex = 0
		c.responseIndex = 1
		c.responseType = ht.Out(1)
	case ht.Out(1) == statusType:
		c.statusIndex = 1
		c.responseIndex = 0
		c.responseType = ht.Out(0)
	case ht.Out(0).Kind() == reflect.Int && ht.Out(1).Kind() == reflect.Int:
		panic(errorMsg)
	case ht.Out(0).Kind() == reflect.Int:
		c.statusIndex = 0
		c.responseIndex = 1
		c.responseType = ht.Out(1)
	case ht.Out(1).Kind() == reflect.Int:
		c.statusIndex = 1
		c.responseIndex = 0
		c.responseType = ht.Out(0)
	default:
		panic(errorMsg)
	}
}

func (c *ErrorHandlerCaller) call(context *callContext) {
	results := c.handler.Call([]reflect.Value{*context.error})

	status := results[c.statusIndex].Convert(intType).Interface().(int)
	response := results[c.responseIndex].Interface()
	if status == 0 {
		status = 500
		response = &DefaultErrorResponse{
			Message: "internal server error",
			Success: false,
		}
	}

	context.rawContext.JSON(status, response)
}
