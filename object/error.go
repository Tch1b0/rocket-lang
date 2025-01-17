package object

import "fmt"

type Error struct {
	Message string
}

func NewError(e interface{}) *Error {
	return &Error{Message: fmt.Sprintf("%s", e)}
}

func NewErrorFormat(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) InvokeMethod(method string, env Environment, args ...Object) Object {
	return objectMethodLookup(e, method, args)
}
