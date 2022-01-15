package docs

import (
	_ "embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
)

//go:embed swagger_template.html
var docsTemplateStr string

type Handler struct {
	docs     *Docs
	template *template.Template
}

func NewHandler(docs *Docs) *Handler {
	tmpl := template.Must(template.New("").Parse(docsTemplateStr))
	return &Handler{docs: docs, template: tmpl}
}

func (h *Handler) Docs(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Status(http.StatusOK)
	err := h.template.Execute(ctx.Writer, h.docs)
	if err != nil {
		panic(err.Error())
	}
}

func (h *Handler) File(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.docs.OpenAPIContent())
}
