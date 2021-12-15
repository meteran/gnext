package docs

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	docs *Docs
}

func NewHandler(docs *Docs) *Handler {
	return &Handler{docs: docs}
}

func (h *Handler) Docs(ctx *gin.Context) {
	ctx.HTML(200, "docs.html", h.docs)
}

func (h *Handler) File(ctx *gin.Context) {
	doc, err := openapi3.NewLoader().LoadFromFile("docs/openapi.json")
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(200, doc)
}