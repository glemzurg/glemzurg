package model_data_type

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

type AtomicEnum struct {
	Value     string
	SortOrder int
}

func (a *AtomicEnum) Validate(ctx *coreerr.ValidationContext) error {
	if a.Value == "" {
		return coreerr.New(ctx, coreerr.DtypeEnumValueRequired, "Value is required", "Value")
	}
	return nil
}
