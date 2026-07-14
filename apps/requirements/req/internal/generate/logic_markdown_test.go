package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
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

	got := logicMarkdownSpecLinesForClass(class, logic, nil, nil)
	require.Equal(t, "    - **self.Amount /= 0**", got)
}

func TestLogicMarkdownSpecLinesAssociationClassReify(t *testing.T) {
	txKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"transaction",
	))
	acKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"account_balance_change",
	))
	acctKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"account",
	))
	assocKey := helper.Must(identity.NewClassAssociationKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		txKey, acctKey, "adjusts",
	))
	assoc := model_class.NewAssociation(
		assocKey,
		model_class.AssociationDetails{Name: "Adjusts", Details: ""},
		model_class.AssociationEnd{ClassKey: txKey, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
		model_class.AssociationEnd{ClassKey: acctKey, Multiplicity: helper.Must(model_class.NewMultiplicity("1..many"))},
		model_class.AssociationOptions{AssociationClassKey: &acKey},
	)
	acClass := model_class.NewClass(acKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account Balance Change"})
	txClass := model_class.NewClass(txKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Transaction"})

	logic := model_logic.NewLogic(
		identity.Key{},
		model_logic.LogicTypeStateChange,
		"The data for each Adjusts.",
		"AccountBalanceChange",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "_new(r.amount)"},
		nil,
	)
	logic.SetEndpointSelectorSpec(logic_spec.ExpressionSpec{
		Notation:      model_logic.NotationTLAPlus,
		Specification: `{ r.account : r \in Amounts }`,
	})

	got := logicMarkdownSpecLinesForClass(
		txClass,
		logic,
		map[identity.Key]model_class.Association{assocKey: assoc},
		map[identity.Key]model_class.Class{acKey: acClass},
	)
	require.Equal(t, strings.Join([]string{
		`    - **Adjusts selector: { r.account : r \in Amounts }**`,
		`    - **AccountBalanceChange' = «new»(r.amount)**`,
	}, "\n"), got)
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

func TestParameterSimulationMarkdownLines(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet")),
		"transaction",
	))
	actionKey := helper.Must(identity.NewActionKey(classKey, "initialize"))
	paramKey := helper.Must(identity.NewParameterKey(actionKey, "amounts"))
	reqKey := helper.Must(identity.NewParameterSimulationRequireKey(paramKey, "0"))
	specKey := helper.Must(identity.NewParameterSimulationSpecKey(paramKey, "0"))

	param := helper.Must(model_state.NewParameter(actionKey, "Amounts", "unordered of unconstrained", false))
	specLogic := model_logic.NewLogic(
		specKey,
		model_logic.LogicTypeValue,
		"",
		"",
		logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "{}"},
		nil,
	)
	param.SetSimulation(&model_state.ParameterSimulation{
		Details: "Sample amounts.",
		Rules: []model_state.ParameterSimulationRule{{
			Requires: []model_logic.Logic{
				model_logic.NewLogic(
					reqKey,
					model_logic.LogicTypeAssessment,
					"Need accounts.",
					"",
					logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "Account /= {}"},
					nil,
				),
			},
			Specification: &specLogic,
		}},
	})
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Transaction"})

	got := parameterSimulationMarkdownLines(class, param)
	require.Contains(t, got, "    - Simulation:")
	require.Contains(t, got, "Sample amounts.")
	require.Contains(t, got, "            - Requires:")
	require.Contains(t, got, "Need accounts.")
	require.Contains(t, got, "**Account /= {}**")
	require.Contains(t, got, "            - Specification:\n                - **{}**")
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
