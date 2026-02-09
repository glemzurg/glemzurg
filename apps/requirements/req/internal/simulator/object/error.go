package object

import "fmt"

// Error represents an error during evaluation.
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return TypeError }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

func (e *Error) SetValue(source Object) error {
	return fmt.Errorf("cannot assign to Error")
}

func (e *Error) Clone() Object {
	return &Error{Message: e.Message}
}
