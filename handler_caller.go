package gnext

import (
	"reflect"
)

func NewHandlerCaller(receiver reflect.Value) *HandlerCaller {
	return &HandlerCaller{
		receiver: receiver,
	}
}

type HandlerCaller struct {
	receiver    reflect.Value
	argBuilders []argBuilder
	argSetters  []argSetter
	errorIndex  int
}

func (c *HandlerCaller) addSetter(setter argSetter) {
	c.argSetters = append(c.argSetters, setter)
}

func (c *HandlerCaller) addBuilder(b argBuilder) {
	c.argBuilders = append(c.argBuilders, b)
}

func (c *HandlerCaller) call(ctx *callContext) error {
	values := make([]reflect.Value, len(c.argBuilders))
	for i, builder := range c.argBuilders {
		value, err := builder(ctx)
		if err != nil {
			ctx.error = err
			return err
		}
		values[i] = value
	}

	results := c.receiver.Call(values)
	for i, setter := range c.argSetters {
		setter(&results[i], ctx)
	}
	return ctx.error
}
