package gnext

import (
	"github.com/gin-gonic/gin"
	"reflect"
)

type callContext struct {
	rawContext *gin.Context
	values     []*reflect.Value
	error      error
	status     Status
}
