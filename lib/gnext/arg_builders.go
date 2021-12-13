package gnext

import (
	"github.com/gin-gonic/gin/binding"
	"io"
	"reflect"
	"strconv"
)

// TODO potential optimization: return a pointer to `reflect.value` instead of struct directly
type argBuilder func(ctx *callContext) (reflect.Value, error)

func stringParamBuilder(paramName string, optional bool) argBuilder {
	if optional {
		return func(ctx *callContext) (reflect.Value, error) {
			param := ctx.rawContext.Param(paramName)
			if param == "" {
				return reflect.ValueOf((*string)(nil)), nil
			}
			return reflect.ValueOf(&param), nil
		}
	} else {
		return func(ctx *callContext) (reflect.Value, error) {
			param := ctx.rawContext.Param(paramName)
			if param == "" {
				return reflect.Value{}, NotFound
			}
			return reflect.ValueOf(param), nil
		}
	}
}

func intParamBuilder(paramName string, optional bool) argBuilder {
	if optional {
		return func(ctx *callContext) (reflect.Value, error) {
			number, err := strconv.Atoi(ctx.rawContext.Param(paramName))
			if err != nil {
				return reflect.ValueOf((*int)(nil)), nil
			}
			return reflect.ValueOf(&number), nil
		}
	} else {
		return func(ctx *callContext) (reflect.Value, error) {
			number, err := strconv.Atoi(ctx.rawContext.Param(paramName))
			if err != nil {
				return reflect.Value{}, NotFound
			}
			return reflect.ValueOf(number), nil
		}
	}
}

func paramBuilder(kind reflect.Kind, paramName string, optional bool) argBuilder {
	switch kind {
	case reflect.Int:
		return intParamBuilder(paramName, optional)
	case reflect.String:
		return stringParamBuilder(paramName, optional)
	default:
		panic("unknown param kind: " + kind.String())
	}
}

func genericBuilder(bodyType reflect.Type, bindType binding.Binding) argBuilder {
	if bodyType.Kind() == reflect.Ptr {
		bodyType = bodyType.Elem()
		return func(ctx *callContext) (reflect.Value, error) {
			value := reflect.New(bodyType)

			if err := ctx.rawContext.ShouldBindWith(value.Interface(), bindType); err != nil {
				if err == io.EOF {
					return reflect.New(value.Type()).Elem(), nil
				}
				return reflect.Value{}, err
			}

			return value, nil
		}
	} else {
		return func(ctx *callContext) (reflect.Value, error) {
			value := reflect.New(bodyType)

			if err := ctx.rawContext.ShouldBindWith(value.Interface(), bindType); err != nil {
				return reflect.Value{}, err
			}

			return value.Elem(), nil
		}
	}
}

func rawContextBuilder(ctx *callContext) (reflect.Value, error) {
	return reflect.ValueOf(ctx.rawContext), nil
}

func cached(builder argBuilder, cacheIndex int) argBuilder {
	return func(ctx *callContext) (reflect.Value, error) {
		value, err := builder(ctx)
		ctx.values[cacheIndex] = &value
		return value, err
	}
}

func cachedValue(cacheIndex int) argBuilder {
	return func(ctx *callContext) (reflect.Value, error) {
		return *ctx.values[cacheIndex], nil
	}
}
