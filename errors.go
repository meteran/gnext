package gnext

import "fmt"

type NotFound struct{ error }

type HandlerPanicked struct {
	Value      interface{}
	StackTrace []byte
}

func (e *HandlerPanicked) Error() string {
	return fmt.Sprintf("%v:\n%s", e.Value, e.StackTrace)
}
