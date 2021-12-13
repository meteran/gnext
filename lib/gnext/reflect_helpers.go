package gnext

import "reflect"

func typesEqual(expected reflect.Type, given reflect.Type) bool {
	if given.Kind() == reflect.Ptr {
		given = given.Elem()
	}
	return given == expected
}
