package gnext

import "reflect"

type argSetter func(*reflect.Value, *callContext)

func valueSetter(contextIndex int) argSetter {
	return func(value *reflect.Value, ctx *callContext) {
		ctx.values[contextIndex] = value
	}
}

func responseSetter(contextIndex int) argSetter {
	return func(value *reflect.Value, ctx *callContext) {
		ctx.values[contextIndex] = value
		ctx.responseIndex = contextIndex
	}
}

func errorSetter(value *reflect.Value, ctx *callContext) {
	if !value.IsNil() {
		ctx.error = value
	}
}

func headersSetter(optional bool) argSetter {
	if optional {
		return func(value *reflect.Value, ctx *callContext) {
			headers := value.Interface().(*Headers)
			if headers != nil {
				for key, values := range *headers {
					for _, val := range values {
						ctx.rawContext.Header(key, val)
					}
				}
			}
		}
	} else {
		return func(value *reflect.Value, ctx *callContext) {
			for key, values := range value.Interface().(Headers) {
				for _, val := range values {
					ctx.rawContext.Header(key, val)
				}
			}
		}
	}
}

func statusSetter(optional bool) argSetter {
	if optional {
		return func(value *reflect.Value, ctx *callContext) {
			statusPtr := value.Interface().(*Status)
			if statusPtr != nil {
				ctx.status = *statusPtr
			}
		}
	} else {
		return func(value *reflect.Value, ctx *callContext) {
			ctx.status = value.Interface().(Status)
		}
	}
}
