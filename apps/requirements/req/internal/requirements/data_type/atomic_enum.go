package data_type

import validation "github.com/go-ozzo/ozzo-validation/v4"

type AtomicEnum struct {
	Value     string
	SortOrder int
}

func (a *AtomicEnum) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Value, validation.Required),
	)
}
