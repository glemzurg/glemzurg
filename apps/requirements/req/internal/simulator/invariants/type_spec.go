package invariants

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
)

func dataTypeHasTypeSpec(dataType *model_data_type.DataType) bool {
	if dataType == nil || dataType.TypeSpec == nil {
		return false
	}
	return strings.TrimSpace(dataType.TypeSpec.Specification) != ""
}

func attributeHasTypeSpec(attr *model_class.Attribute) bool {
	if attr == nil {
		return false
	}
	return dataTypeHasTypeSpec(attr.DataType)
}
