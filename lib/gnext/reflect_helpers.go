package gnext

import "reflect"

func typesEqual(expected reflect.Type, given reflect.Type) bool {
	if isPtr(given) {
		given = given.Elem()
	}
	return given == expected
}

func isPtr(arg reflect.Type) bool {
	return arg.Kind() == reflect.Ptr
}
