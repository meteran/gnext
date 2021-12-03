package gnext

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"reflect"
	"strconv"
)

type builder func(ctx *gin.Context) (reflect.Value, error)

func stringParamBuilder(paramName string) builder {
	return func(ctx *gin.Context) (reflect.Value, error) {
		param := ctx.Param(paramName)
		if param == "" {
			return reflect.Value{}, NotFound
		}
		return reflect.ValueOf(param), nil
	}
}

func intParamBuilder(paramName string) builder {
	return func(ctx *gin.Context) (reflect.Value, error) {
		number, err := strconv.Atoi(ctx.Param(paramName))
		if err != nil {
			return reflect.Value{}, NotFound
		}
		return reflect.ValueOf(number), nil
	}
}

func stringOptionalParamBuilder(paramName string) builder {
	return func(ctx *gin.Context) (reflect.Value, error) {
		param := ctx.Param(paramName)
		if param == "" {
			return reflect.ValueOf((*string)(nil)), nil
		}
		return reflect.ValueOf(&param), nil
	}
}

func intOptionalParamBuilder(paramName string) builder {
	return func(ctx *gin.Context) (reflect.Value, error) {
		number, err := strconv.Atoi(ctx.Param(paramName))
		if err != nil {
			return reflect.ValueOf((*int)(nil)), nil
		}
		return reflect.ValueOf(&number), nil
	}
}

func headerBuilder(optional bool) builder {
	return func(ctx *gin.Context) (reflect.Value, error) {
		h := Headers(ctx.Request.Header)
		if optional {
			return reflect.ValueOf(&h), nil
		}
		return reflect.ValueOf(h), nil
	}
}

func genericBuilder(bodyType reflect.Type, bindType binding.Binding, optional bool) builder {
	return func(ctx *gin.Context) (reflect.Value, error) {
		value := reflect.New(bodyType)

		if err := ctx.ShouldBindWith(value.Interface(), bindType); err != nil {
			if err == io.EOF && optional {
				return reflect.New(value.Type()).Elem(), nil
			}
			return reflect.Value{}, err
		}

		if optional {
			return value, nil
		}
		return value.Elem(), nil
	}
}

