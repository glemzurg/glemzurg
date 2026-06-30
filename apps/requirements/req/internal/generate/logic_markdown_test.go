package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
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

func TestLogicMarkdownSpecLinesForClassRewritesSelfFields(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"account_balance_change",
	))
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeAssessment,
		"Amount cannot be zero.",
		"",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "self.amount /= 0"},
		nil,
	)
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account Balance Change"})
	class.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(
			helper.Must(identity.NewAttributeKey(classKey, "amount")),
			model_class.AttributeDetails{Name: "Amount", Details: ""},
			"",
			nil,
			false,
			model_class.AttributeAnnotations{},
		)),
	})

	got := logicMarkdownSpecLinesForClass(class, logic)
	require.Equal(t, "    - **self.Amount /= 0**", got)
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

func TestLogicMarkdownSpecLinesBoldsDestroyGuarantee(t *testing.T) {
	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeDestroy,
		"Peer _destroy events for removed peers",
		"AppliesSocialCurrencyLogic",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: `{ b \in AppliesSocialCurrencyLogic : TRUE }`},
		nil,
	)
	logic.SetDestroyEventSpec(logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "_destroy(b)"})

	got := logicMarkdownSpecLines(logic)
	require.Equal(t, strings.Join([]string{
		"    - **AppliesSocialCurrencyLogic' = { b \\in AppliesSocialCurrencyLogic : TRUE }**",
		"    - Each removed element sent: **«destroy»(b)**",
	}, "\n"), got)
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
