package invariants

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/stretchr/testify/suite"
)

type NullableWhenSetSuite struct {
	suite.Suite
}

func TestNullableWhenSetSuite(t *testing.T) {
	suite.Run(t, new(NullableWhenSetSuite))
}

func (s *NullableWhenSetSuite) TestNullableWhenSetSpecification() {
	s.Equal("IF ISO = NULL THEN TRUE ELSE ISO \\in _Iso4217Codes", NullableWhenSetSpecification("ISO", "ISO \\in _Iso4217Codes"))
}

func (s *NullableWhenSetSuite) TestLogicSpecHasNullableWhenUnsetGuardDetectsAuthorGuard() {
	pf := convert.NewExpressionParseFunc(nil)
	spec := helper.Must(logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, "IF accountId = NULL THEN TRUE ELSE accountId > 0", pf))
	s.True(LogicSpecHasNullableWhenUnsetGuard(spec))
}

func (s *NullableWhenSetSuite) TestLogicSpecHasNullableWhenUnsetGuardRejectsBareInvariant() {
	pf := convert.NewExpressionParseFunc(nil)
	spec := helper.Must(logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, "accountId > 0", pf))
	s.False(LogicSpecHasNullableWhenUnsetGuard(spec))
}
