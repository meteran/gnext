package gnext

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type handler struct {
	called bool
}

func (h *handler) call() {
	h.called = true
}

func TestErrorFallbacksToMiddleware(t *testing.T) {
	r := Router()

	middlewares := make([]handler, 7)
	testHandler := handler{}

	var (
		errorFlag bool
	)

	r.OnError(func(err error) (Status, interface{}) {
		errorFlag = true
		return 422, nil
	})

	r.Use(Middleware{
		Before: middlewares[0].call,
	})

	r.Use(Middleware{
		After: middlewares[1].call,
	})

	r.Use(Middleware{
		Before: middlewares[2].call,
		After:  middlewares[3].call,
	})

	r.Use(Middleware{
		Before: func() error {
			return fmt.Errorf("error")
		},
		After: middlewares[4].call,
	})

	r.Use(Middleware{
		Before: middlewares[5].call,
	})

	r.Use(Middleware{
		After: middlewares[6].call,
	})

	r.GET("/path", testHandler.call)
	response := makeRequest(t, r, http.MethodGet, "/path")

	assert.True(t, errorFlag)
	assert.True(t, middlewares[0].called)
	assert.True(t, middlewares[1].called)
	assert.True(t, middlewares[2].called)
	assert.True(t, middlewares[3].called)
	assert.True(t, middlewares[4].called)
	assert.False(t, middlewares[5].called)
	assert.False(t, middlewares[6].called)
	assert.False(t, testHandler.called)

	assert.Equal(t, 422, response.Code)
	assert.Equal(t, "null", response.Body.String())
}

func TestOverrideReturnedResponseAndStatusInErrorHandler(t *testing.T) {
	r := Router()

	type responseType string

	r.OnError(func(err error) (responseType, Status) {
		return "overridden", 422
	})

	r.GET("/path", func() (responseType, Status, error) {
		return "from handler", 200, fmt.Errorf("some error")
	})

	response := makeRequest(t, r, http.MethodGet, "/path")
	assert.Equal(t, 422, response.Code)
	assert.Equal(t, `"overridden"`, response.Body.String())
}

func TestOverrideReturnedResponseAndStatusInErrorHandlerAndMiddleware(t *testing.T) {
	r := Router()

	type responseType string

	r.OnError(func(err error) (responseType, Status) {
		return "from error", 422
	})

	r.Use(Middleware{
		After: func() (responseType, Status) {
			return "overridden", 401
		},
	}).GET("/path", func() (responseType, Status, error) {
		return "from handler", 200, fmt.Errorf("some error")
	})

	response := makeRequest(t, r, http.MethodGet, "/path")
	assert.Equal(t, 401, response.Code)
	assert.Equal(t, `"overridden"`, response.Body.String())
}

func TestOverrideReturnedResponseAndStatusFromErrorHandlerInAfterMiddleware(t *testing.T) {
	r := Router()

	type responseType string

	r.OnError(func(err error) (responseType, Status) {
		return "from error", 422
	})

	r.Use(Middleware{
		After: func() (responseType, Status) {
			return "overridden", 401
		},
	}).GET("/path", func() error {
		return fmt.Errorf("some error")
	})

	response := makeRequest(t, r, http.MethodGet, "/path")
	assert.Equal(t, 401, response.Code)
	assert.Equal(t, `"overridden"`, response.Body.String())
}

func TestOverrideReturnedResponseAndStatusEvenForDifferentTypes(t *testing.T) {
	r := Router()

	type responseType string
	type overriddenType int

	r.OnError(func(err error) (responseType, Status) {
		return "from error", 422
	})

	r.Use(Middleware{
		After: func() (overriddenType, Status) {
			return 10, 401
		},
	}).GET("/path", func() (overriddenType, error) {
		return 5, fmt.Errorf("some error")
	})

	response := makeRequest(t, r, http.MethodGet, "/path")
	assert.Equal(t, 401, response.Code)
	assert.Equal(t, `10`, response.Body.String())
}

func TestOverrideReturnedResponseAndStatusAfterErrorAndHandlerWithoutResponse(t *testing.T) {
	r := Router()

	type responseType string
	type overriddenType struct {
		Response
		Value string `json:"value"`
	}

	r.OnError(func(err error) (responseType, Status) {
		return "from error", 422
	})

	r.Use(Middleware{
		After: func() (overriddenType, Status) {
			return overriddenType{Value: "overridden"}, 401
		},
	}).GET("/path", func() error {
		return fmt.Errorf("some error")
	})

	response := makeRequest(t, r, http.MethodGet, "/path")
	assert.Equal(t, 401, response.Code)
	assert.Equal(t, `{"value":"overridden"}`, response.Body.String())
}
