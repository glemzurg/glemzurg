package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestLogicMarkdownSpecLinesBoldsAssessmentSpec(t *testing.T) {
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeAssessment,
		"Balance cannot be negative.",
		"",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "self.balance >= 0"},
		nil,
	)

	got := logicMarkdownSpecLines(logic)
	require.Equal(t, "    - **self.balance >= 0**", got)
}

func TestLogicMarkdownSpecLinesBoldsStateChangeAssignment(t *testing.T) {
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeStateChange,
		"Set name.",
		"name",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Name"},
		nil,
	)

	got := logicMarkdownSpecLines(logic)
	require.Equal(t, "    - **name' = Name**", got)
}

func TestLogicMarkdownSpecLinesBoldsLetBinding(t *testing.T) {
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeLet,
		"Bind helper.",
		"total",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "self.price + self.tax"},
		nil,
	)

	got := logicMarkdownSpecLines(logic)
	require.Equal(t, "    - **LET total = self.price + self.tax**", got)
}

func TestDerivationPolicyMarkdownHTMLBoldsSpec(t *testing.T) {
	policy := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeValue,
		"Derived from sets.",
		"",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "_Bags!SetToBag(set1)"},
		nil,
	)

	got := derivationPolicyMarkdownHTML(&policy)
	require.Equal(t, "Derived from sets.<br>**_Bags!SetToBag(set1)**", got)
}
