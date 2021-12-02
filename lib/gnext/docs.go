package gnext

import (
	"github.com/gin-gonic/gin"
)

func NewDocs() *Docs {
	return &Docs{}
}

type Docs struct {
	content []interface{}
}

func (d *Docs) handler(ctx *gin.Context) {
	ctx.JSON(200, d.content)
}

func (d *Docs) append(docs interface{}) {
	d.content = append(d.content, docs)
}
