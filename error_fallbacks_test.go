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
		return 500, nil
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
	makeRequest(t, r, http.MethodGet, "/path")

	assert.True(t, errorFlag)
	assert.True(t, middlewares[0].called)
	assert.True(t, middlewares[1].called)
	assert.True(t, middlewares[2].called)
	assert.True(t, middlewares[3].called)
	assert.True(t, middlewares[4].called)
	assert.False(t, middlewares[5].called)
	assert.False(t, middlewares[6].called)
	assert.False(t, testHandler.called)
}
