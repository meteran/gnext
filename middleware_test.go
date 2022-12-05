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

func TestBeforeMiddlewareWithParamInHeaders(t *testing.T) {
	h := &handler{}

	type middlewareHeaders struct {
		Headers
		ContentType string `header:"Content-Type"`
	}

	middleware := func() Middleware {
		return Middleware{
			Before: func(headers *middlewareHeaders) error {
				if headers.ContentType != "some content type" {
					return fmt.Errorf("unexpected content type")
				}
				return nil
			},
		}
	}

	r := Router()
	r.Use(middleware())
	r.GET("/foo", func() {
		h.called = true
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 500)
	assert.False(t, h.called)
}
func TestBeforeMiddlewareReturnSomeContext(t *testing.T) {
	h := &handler{}

	type someMiddlewareContext struct {
		SomeContextValue string
	}

	middleware := func() Middleware {
		return Middleware{
			Before: func() *someMiddlewareContext {
				return &someMiddlewareContext{SomeContextValue: "test value"}
			},
		}
	}

	r := Router()
	r.Use(middleware())
	r.GET("/foo", func(someMiddlewareCtx *someMiddlewareContext) {
		h.called = true
		assert.Equal(t, "test value", someMiddlewareCtx.SomeContextValue)
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, 200, res.Code)
	assert.True(t, h.called)
}
func TestBeforeMiddlewareUsingOtherBeforeMiddlewareContext(t *testing.T) {
	h := &handler{}

	type firstMiddlewareContext struct {
		SomeContextValue string
	}

	type secondMiddlewareContext struct {
		SomeContextValue string
	}

	firstMiddleware := func() Middleware {
		return Middleware{
			Before: func() *firstMiddlewareContext {
				return &firstMiddlewareContext{SomeContextValue: "test value"}
			},
		}
	}

	secondMiddleware := func() Middleware {
		return Middleware{
			Before: func(firstMiddlewareCtx *firstMiddlewareContext) *secondMiddlewareContext {
				return &secondMiddlewareContext{SomeContextValue: firstMiddlewareCtx.SomeContextValue}
			},
		}
	}

	r := Router()
	r.Use(firstMiddleware())
	r.Use(secondMiddleware())
	r.GET("/foo", func(firstMiddlewareCtx *firstMiddlewareContext, secondMiddlewareCtx *secondMiddlewareContext) {
		h.called = true
		assert.Equal(t, "test value", firstMiddlewareCtx.SomeContextValue)
		assert.Equal(t, "test value", secondMiddlewareCtx.SomeContextValue)
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, 200, res.Code)
	assert.True(t, h.called)
}
func TestAfterMiddlewareUsingSomeTypeFromHandler(t *testing.T) {
	h := &handler{}

	type response struct {
		Response
		Message string
	}

	middleware := func() Middleware {
		return Middleware{
			After: func(res *response) {
				assert.Equal(t, "test", res.Message)
			},
		}
	}

	r := Router()
	r.Use(middleware())
	r.GET("/foo", func() *response {
		h.called = true
		return &response{Message: "test"}
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 200)
	assert.True(t, h.called)
}
func TestPassBeforeMiddlewareContextToAfterMiddleware(t *testing.T) {
	h := &handler{}

	type someMiddlewareContext struct {
		SomeContextValue string
	}

	middleware := func() Middleware {
		return Middleware{
			Before: func() *someMiddlewareContext {
				return &someMiddlewareContext{SomeContextValue: "test"}
			},
			After: func(someMiddlewareCtx *someMiddlewareContext) {
				assert.Equal(t, "test", someMiddlewareCtx.SomeContextValue)
			},
		}
	}

	r := Router()
	r.Use(middleware())
	r.GET("/foo", func(someMiddleWareCtx *someMiddlewareContext) *someMiddlewareContext {
		h.called = true
		return someMiddleWareCtx
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 200)
	assert.True(t, h.called)
}
func TestAfterMiddlewareUsingOtherAfterMiddlewareContext(t *testing.T) {
	h := &handler{}

	type someMiddlewareContext struct {
		SomeContextValue string
	}

	firstMiddleware := func() Middleware {
		return Middleware{
			After: func(someMiddlewareCtx *someMiddlewareContext) *someMiddlewareContext {
				return someMiddlewareCtx
			},
		}
	}

	secondMiddleware := func() Middleware {
		return Middleware{
			After: func(someMiddlewareCtx *someMiddlewareContext) {
				assert.Equal(t, "test", someMiddlewareCtx.SomeContextValue)
			},
		}
	}

	r := Router()
	r.Use(firstMiddleware())
	r.Use(secondMiddleware())
	r.GET("/foo", func() *someMiddlewareContext {
		h.called = true
		return &someMiddlewareContext{SomeContextValue: "test"}
	})

	res := makeRequest(t, r, http.MethodGet, "/foo")

	assert.Equal(t, res.Code, 200)
	assert.True(t, h.called)
}
