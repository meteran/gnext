package gnext

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

//go:embed test_data/RouteErrorsToSpecificHandlersExpectedDocs.json
var routeErrorsToSpecificHandlersExpectedDocs string

func TestRouteErrorsToSpecificHandlers(t *testing.T) {
	r := Router()

	type globalError struct{ error }
	type specificError struct{ error }
	type unknownError struct{ error }

	type globalResponse struct {
		ErrorResponse `default_status:"501" status_codes:"502"`
		Message       string `json:"message"`
	}

	type specificResponse struct {
		ErrorResponse `default_status:"400" status_codes:"401,403"`
		Code          int `json:"code"`
	}

	type overwritingResponse struct {
		ErrorResponse `default_status:"422"`
		Error         string `json:"error"`
	}

	fallbackResp := "fallback"
	globalResp := &globalResponse{Message: "global"}
	specificResp := &specificResponse{Code: 10}
	overwritingResp := &overwritingResponse{Error: "overwriting"}

	// error handlers
	fallbackHandler := func(err error) (string, Status) {
		return fallbackResp, 404
	}

	globalHandler := func(err *globalError) *globalResponse {
		return globalResp
	}

	specificHandler := func(err *specificError) (*specificResponse, Status) {
		return specificResp, 401
	}

	overwritingHandler := func(err specificError) *overwritingResponse {
		return overwritingResp
	}

	// handlers
	handlerRaisingUnknownError := func() (interface{}, error) {
		return nil, &unknownError{fmt.Errorf("unknown")}
	}

	handlerRaisingGlobalError := func() (interface{}, error) {
		return nil, &globalError{fmt.Errorf("globalError")}
	}

	handlerRaisingSpecificError := func() (interface{}, error) {
		return nil, &specificError{fmt.Errorf("specificError")}
	}

	// routes
	r.GET("/default-unknown", handlerRaisingUnknownError)
	r.OnError(globalHandler)
	r.OnError(fallbackHandler)
	r.GET("/fallback-unknown", handlerRaisingUnknownError)
	r.GET("/global-global", handlerRaisingGlobalError)
	r.GET("/fallback-specific", handlerRaisingSpecificError)
	r.OnError(specificHandler)
	r.GET("/specific-specific", handlerRaisingSpecificError)
	group := r.Group("/grouped")
	group.GET("/fallback-unknown", handlerRaisingUnknownError)
	group.GET("/global-global", handlerRaisingGlobalError)
	group.GET("/specific-specific", handlerRaisingSpecificError)
	group.OnError(overwritingHandler)
	group.GET("/overwriting-specific", handlerRaisingSpecificError)

	docs, err := json.Marshal(r.Docs.OpenApi)
	require.NoError(t, err)
	assert.JSONEq(t, routeErrorsToSpecificHandlersExpectedDocs, string(docs))

	cases := []struct {
		path     string
		status   int
		response string
	}{
		{
			path:     "/default-unknown",
			status:   500,
			response: `{"details": null, "message": "internal server error", "success": false}`,
		},
		{
			path:     "/fallback-unknown",
			status:   404,
			response: `"fallback"`,
		},
		{
			path:     "/global-global",
			status:   501,
			response: `{"message":"global"}`,
		},
		{
			path:     "/fallback-specific",
			status:   404,
			response: `"fallback"`,
		},
		{
			path:     "/specific-specific",
			status:   401,
			response: `{"code": 10}`,
		},
		{
			path:     "/grouped/fallback-unknown",
			status:   0,
			response: `{}`,
		},
		{
			path:     "/grouped/global-global",
			status:   0,
			response: `{}`,
		},
		{
			path:     "/grouped/specific-specific",
			status:   0,
			response: `{}`,
		},
		{
			path:     "/grouped/overwriting-specific",
			status:   0,
			response: `{}`,
		},
	}

	for idx, c := range cases {
		response := makeRequest(t, r, "GET", c.path)
		assert.Equalf(t, c.status, response.Code, "case: %d, path: %s", idx, c.path)
		assert.JSONEqf(t, c.response, response.Body.String(), "case: %d, path: %s", idx, c.path)
	}
}
