package model_data_type

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

type AtomicEnum struct {
	Value     string
	SortOrder int
}

func (a *AtomicEnum) Validate() error {
	if a.Value == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.DtypeEnumValueRequired,
			Message: "Value is required",
			Field:   "Value",
		}
	}
	return nil
}
