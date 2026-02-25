package model_data_type

type AtomicEnum struct {
	Value     string `validate:"required"`
	SortOrder int
}

func (a *AtomicEnum) Validate() error {
	return _validate.Struct(a)
}
