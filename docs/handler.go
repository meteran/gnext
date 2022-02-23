package docs

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"html/template"
	"net/http"
)

//go:embed swagger_template.html
var docsTemplateStr string

type Handler struct {
	docs     *Docs
	template *template.Template
	port     string
}

func NewHandler(docs *Docs, port string) *Handler {
	tmpl := template.Must(template.New("").Parse(docsTemplateStr))
	return &Handler{docs: docs, template: tmpl, port: port}
}

func (h *Handler) Docs(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Status(http.StatusOK)

	values := map[string]interface{}{
		"title":   h.docs.OpenApi.Info.Title,
		"jsonUrl": fmt.Sprintf("http://localhost:%s%s", h.port, h.docs.JsonUrl),
	}

	err := h.template.Execute(ctx.Writer, values)
	if err != nil {
		panic(err.Error())
	}
}

func (h *Handler) JsonFile(ctx *gin.Context) {
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", h.docs.OpenApi.Info.Title))
	ctx.JSON(http.StatusOK, h.docs.OpenApi)
}

func (h *Handler) YamlFile(ctx *gin.Context) {
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.yaml", h.docs.OpenApi.Info.Title))

	bytes, err := yaml.Marshal(h.docs.OpenApi)
	if err != nil {
		panic(err)
	}

	ctx.Data(http.StatusOK, "application/x-yaml; charset=utf-8", bytes)
}
