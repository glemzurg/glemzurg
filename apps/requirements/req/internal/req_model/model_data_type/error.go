package model_data_type

import "fmt"

type CannotParseError struct {
	err   error  // Wrapped original error
	input string // The input that caused the error
}

// Error satisfies the error interface.
func (e *CannotParseError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("failed to parse '%s': %v", e.input, e.err)
	}
	return fmt.Sprintf("failed to parse '%s'", e.input)
}

// Unwrap allows access to the inner error (enables errors.Unwrap, errors.Is, errors.As).
func (e *CannotParseError) Unwrap() error {
	return e.err
}
