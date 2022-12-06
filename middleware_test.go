package gnext

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBeforeMiddlewareWithParamInHeaders(t *testing.T) {
	var called bool

	type middlewareHeaders struct {
		Headers
		ContentType string `header:"Content-Type"`
	}

	middleware := Middleware{
		Before: func(headers *middlewareHeaders) error {
			if headers.ContentType != "some content type" {
				return fmt.Errorf("unexpected content type")
			}
			return nil
		},
	}

	r := Router()
	r.Use(middleware)
	r.GET("/foo", func() {
		called = true
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 500)
	assert.False(t, called)
}
func TestBeforeMiddlewareReturnSomeContext(t *testing.T) {
	var called bool

	type someMiddlewareContext struct {
		SomeContextValue string
	}

	middleware := Middleware{
		Before: func() *someMiddlewareContext {
			return &someMiddlewareContext{SomeContextValue: "test value"}
		},
	}

	r := Router()
	r.Use(middleware)
	r.GET("/foo", func(someMiddlewareCtx *someMiddlewareContext) {
		called = true
		assert.Equal(t, "test value", someMiddlewareCtx.SomeContextValue)
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, 200, res.Code)
	assert.True(t, called)
}
func TestBeforeMiddlewareUsingOtherBeforeMiddlewareContext(t *testing.T) {
	var called bool

	type firstMiddlewareContext struct {
		SomeContextValue string
	}

	type secondMiddlewareContext struct {
		SomeContextValue string
	}

	firstMiddleware := Middleware{
		Before: func() *firstMiddlewareContext {
			return &firstMiddlewareContext{SomeContextValue: "test value"}
		},
	}

	secondMiddleware := Middleware{
		Before: func(firstMiddlewareCtx *firstMiddlewareContext) (*firstMiddlewareContext, *secondMiddlewareContext) {
			firstMiddlewareCtx.SomeContextValue = "changed test value"
			return firstMiddlewareCtx, &secondMiddlewareContext{SomeContextValue: firstMiddlewareCtx.SomeContextValue}
		},
	}

	thirdMiddleware := Middleware{
		Before: func(firstMiddlewareCtx *firstMiddlewareContext, secondMiddlewareCtx *secondMiddlewareContext) {
			assert.Equal(t, "changed test value", firstMiddlewareCtx.SomeContextValue)
			assert.Equal(t, "changed test value", secondMiddlewareCtx.SomeContextValue)
		},
	}

	r := Router()
	r.Use(firstMiddleware)
	r.Use(secondMiddleware)
	r.Use(thirdMiddleware)
	r.GET("/foo", func(firstMiddlewareCtx *firstMiddlewareContext, secondMiddlewareCtx *secondMiddlewareContext) {
		called = true
		assert.Equal(t, "changed test value", firstMiddlewareCtx.SomeContextValue)
		assert.Equal(t, "changed test value", secondMiddlewareCtx.SomeContextValue)
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, 200, res.Code)
	assert.True(t, called)
}
func TestAfterMiddlewareUsingSomeTypeFromHandler(t *testing.T) {
	var called bool

	type response struct {
		Response
		Message string
	}

	middleware := Middleware{
		After: func(res *response) {
			assert.Equal(t, "test", res.Message)
		},
	}

	r := Router()
	r.Use(middleware)
	r.GET("/foo", func() *response {
		called = true
		return &response{Message: "test"}
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 200)
	assert.True(t, called)
}
func TestPassBeforeMiddlewareContextToAfterMiddleware(t *testing.T) {
	var called bool

	type someMiddlewareContext struct {
		SomeContextValue string
	}

	middleware := Middleware{
		Before: func() *someMiddlewareContext {
			return &someMiddlewareContext{SomeContextValue: "test"}
		},
		After: func(someMiddlewareCtx *someMiddlewareContext) {
			assert.Equal(t, "test", someMiddlewareCtx.SomeContextValue)
		},
	}

	r := Router()
	r.Use(middleware)
	r.GET("/foo", func(someMiddleWareCtx *someMiddlewareContext) *someMiddlewareContext {
		called = true
		return someMiddleWareCtx
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 200)
	assert.True(t, called)
}
func TestAfterMiddlewareUsingOtherAfterMiddlewareContext(t *testing.T) {
	var called bool

	type firstMiddlewareContext struct {
		SomeContextValue string
	}

	type secondMiddlewareContext struct {
		SomeContextValue string
	}

	firstMiddleware := Middleware{
		After: func(firstMiddlewareCtx *firstMiddlewareContext) *firstMiddlewareContext {
			return firstMiddlewareCtx
		},
	}

	secondMiddleware := Middleware{
		After: func(firstMiddlewareCtx *firstMiddlewareContext) (*firstMiddlewareContext, *secondMiddlewareContext) {
			assert.Equal(t, "test", firstMiddlewareCtx.SomeContextValue)

			firstMiddlewareCtx.SomeContextValue = "changed test value"
			return firstMiddlewareCtx, &secondMiddlewareContext{SomeContextValue: firstMiddlewareCtx.SomeContextValue}
		},
	}

	thirdMiddleware := Middleware{
		After: func(firstMiddlewareCtx *firstMiddlewareContext, secondMiddlewareContext *secondMiddlewareContext) {
			assert.Equal(t, "changed test value", firstMiddlewareCtx.SomeContextValue)
			assert.Equal(t, "changed test value", secondMiddlewareContext.SomeContextValue)
		},
	}

	r := Router()
	r.Use(firstMiddleware)
	r.Use(secondMiddleware)
	r.Use(thirdMiddleware)
	r.GET("/foo", func() *firstMiddlewareContext {
		called = true
		return &firstMiddlewareContext{SomeContextValue: "test"}
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 200)
	assert.True(t, called)
}
