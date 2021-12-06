package gnext

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"reflect"
	"strconv"
)

type builder func(ctx *gin.Context, contextValues []*reflect.Value) (reflect.Value, error)

func stringParamBuilder(paramName string, optional bool) builder {
	if optional {
		return func(ctx *gin.Context, _ []*reflect.Value) (reflect.Value, error) {
			param := ctx.Param(paramName)
			if param == "" {
				return reflect.ValueOf((*string)(nil)), nil
			}
			return reflect.ValueOf(&param), nil
		}
	} else {
		return func(ctx *gin.Context, _ []*reflect.Value) (reflect.Value, error) {
			param := ctx.Param(paramName)
			if param == "" {
				return reflect.Value{}, NotFound
			}
			return reflect.ValueOf(param), nil
		}
	}
}

func intParamBuilder(paramName string, optional bool) builder {
	if optional {
		return func(ctx *gin.Context, _ []*reflect.Value) (reflect.Value, error) {
			number, err := strconv.Atoi(ctx.Param(paramName))
			if err != nil {
				return reflect.ValueOf((*int)(nil)), nil
			}
			return reflect.ValueOf(&number), nil
		}
	} else {
		return func(ctx *gin.Context, _ []*reflect.Value) (reflect.Value, error) {
			number, err := strconv.Atoi(ctx.Param(paramName))
			if err != nil {
				return reflect.Value{}, NotFound
			}
			return reflect.ValueOf(number), nil
		}
	}
}

func paramBuilder(kind reflect.Kind, paramName string, optional bool) builder {
	switch kind {
	case reflect.Int:
		return intParamBuilder(paramName, optional)
	case reflect.String:
		return stringParamBuilder(paramName, optional)
	default:
		panic("unknown param kind: " + kind.String())
	}
}

func headerBuilder(optional bool) builder {
	return func(ctx *gin.Context, _ []*reflect.Value) (reflect.Value, error) {
		h := Headers(ctx.Request.Header)
		if optional {
			return reflect.ValueOf(&h), nil
		}
		return reflect.ValueOf(h), nil
	}
}

func genericBuilder(bodyType reflect.Type, bindType binding.Binding, optional bool) builder {
	return func(ctx *gin.Context, contextValues []*reflect.Value) (reflect.Value, error) {
		value := reflect.New(bodyType)

		if err := ctx.ShouldBindWith(value.Interface(), bindType); err != nil {
			if err == io.EOF && optional {
				return cached(reflect.New(value.Type()).Elem(), contextValues, cacheIndex), nil
			}
			return reflect.Value{}, err
		}

		if optional {
			return value, nil
		}
		return value.Elem(), nil
	}
}

func cached(value reflect.Value, contextValues []*reflect.Value, cacheIndex int) reflect.Value {
	contextValues[cacheIndex] = &value
	return value
}

func cachedBuilder(cacheIndex int) builder {
	return func(ctx *gin.Context, contextValues []*reflect.Value) (reflect.Value, error) {
		return *contextValues[cacheIndex], nil
	}
}
