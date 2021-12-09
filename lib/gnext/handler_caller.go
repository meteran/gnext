package gnext

import "reflect"

type HandlerCaller struct {
	argBuilders []argBuilder
	//argSetters  []argSetter
	receiver    reflect.Value
}
//func (c *HandlerCaller) addSetter(outputIndex, contextIndex int) {
//	setter := func(output []reflect.Value, context []*reflect.Value) {
//		context[contextIndex] = &output[outputIndex]
//	}
//	c.argSetters = append(c.argSetters, setter)
//}

func (c *HandlerCaller) addBuilder(b argBuilder) {
	c.argBuilders = append(c.argBuilders, b)
}