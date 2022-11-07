package gnext

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type option func(recorder *httptest.ResponseRecorder) *httptest.ResponseRecorder

func makeRequest(t *testing.T, router *RootRouter, method, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
