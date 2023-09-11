package gnext

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type option func(recorder *httptest.ResponseRecorder) *httptest.ResponseRecorder

func makeRequest(t *testing.T, router *RootRouter, method, url string, body ...interface{}) *httptest.ResponseRecorder {
	var payload io.Reader = nil

	if len(body) > 0 {
		payloadBytes, err := json.Marshal(body[0])
		require.NoError(t, err)
		payload = bytes.NewReader(payloadBytes)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, payload)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
