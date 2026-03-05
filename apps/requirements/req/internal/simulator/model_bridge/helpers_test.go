package model_bridge

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"
)

// parsedSpec creates a TLA+ ExpressionSpec with the expression parsed via the convert pipeline.
func parsedSpec(tla string) model_spec.ExpressionSpec {
	pf := convert.NewExpressionParseFunc(nil)
	spec := helper.Must(model_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}
