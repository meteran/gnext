package gnext

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPayloadValidation(t *testing.T) {
	type request struct {
		Id   int    `json:"id" binding:"required"`
		Name string `json:"name"`
	}

	type response struct {
		Result string `json:"result"`
	}

	handler := func(req *request) *response {
		return &response{Result: req.Name}
	}

	r := Router()
	r.POST("/handler", handler)

	res := makeRequest(t, r, "POST", "/handler", gin.H{"name": "some name"})
	assert.Equal(t, http.StatusBadRequest, res.Code)
}
