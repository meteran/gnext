package docs

import (
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
	ctx.HTML(http.StatusOK, "docs.html", h.docs)
}

func (h *Handler) File(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.docs.OpenAPIContent())
}
