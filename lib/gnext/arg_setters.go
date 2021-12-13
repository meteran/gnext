package gnext

import "reflect"

type argSetter func(*reflect.Value, *callContext)

func valueSetter(contextIndex int) argSetter {
	return func(value *reflect.Value, ctx *callContext) {
		ctx.values[contextIndex] = value
	}
}

func errorSetter(optional bool) argSetter {
	if optional {
		return func(value *reflect.Value, ctx *callContext) {
			err := value.Interface().(*error)
			if err != nil {
				ctx.error = *err
			}
		}
	} else {
		return func(value *reflect.Value, ctx *callContext) {
			ctx.error = value.Interface().(error)
		}
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
				ctx.status = int(*statusPtr)
			}
		}
	} else {
		return func(value *reflect.Value, ctx *callContext) {
			ctx.status = int(value.Interface().(Status))
		}
	}
}
