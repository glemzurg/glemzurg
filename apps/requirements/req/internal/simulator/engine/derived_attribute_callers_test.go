package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"math/big"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalDerivedAttributes_ExcludesSimulatableCaller(t *testing.T) {
	accountKey := mustKey("domain/finance/subdomain/wallet/class/account")
	ledgerKey := mustKey("domain/finance/subdomain/wallet/class/ledger_entry")
	balanceAttrKey := helper.Must(identity.NewAttributeKey(accountKey, "balance"))
	amountAttrKey := helper.Must(identity.NewAttributeKey(ledgerKey, "amount"))

	balanceDeriv := model_logic.NewLogic(
		mustKey("invariant/10"),
		model_logic.LogicTypeValue,
		"Sum adjustments.",
		"",
		logicSpecWithExpression(&me.IntLiteral{Value: big.NewInt(0)}),
		nil,
	)
	balanceAttr := helper.Must(model_class.NewAttribute(
		balanceAttrKey,
		model_class.AttributeDetails{Name: "balance", Details: ""},
		"",
		&balanceDeriv,
		false,
		model_class.AttributeAnnotations{},
	))

	ledgerStateKey := helper.Must(identity.NewStateKey(ledgerKey, "posted"))
	ledgerActionKey := helper.Must(identity.NewActionKey(ledgerKey, "post"))
	requires := model_logic.NewLogic(
		mustKey("invariant/11"),
		model_logic.LogicTypeAssessment,
		"",
		"",
		logicSpecWithExpression(&me.AttributeRef{AttributeKey: balanceAttrKey}),
		nil,
	)
	ledgerAction := model_state.NewAction(
		ledgerActionKey,
		model_state.ActionDetails{Name: "Post", Details: ""},
		[]model_logic.Logic{requires},
		nil,
		nil,
		nil,
	)

	accountClass := model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"})
	accountClass.SetAttributes([]model_class.Attribute{balanceAttr})
	accountClass.SetStates(map[identity.Key]model_state.State{})
	accountClass.SetEvents(map[identity.Key]model_state.Event{})
	accountClass.SetGuards(map[identity.Key]model_state.Guard{})
	accountClass.SetActions(map[identity.Key]model_state.Action{})
	accountClass.SetQueries(map[identity.Key]model_state.Query{})
	accountClass.SetTransitions(map[identity.Key]model_state.Transition{})

	ledgerClass := model_class.NewClass(ledgerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "LedgerEntry"})
	ledgerClass.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(amountAttrKey, model_class.AttributeDetails{Name: "amount", Details: ""}, "", nil, false, model_class.AttributeAnnotations{})),
	})
	ledgerClass.SetStates(map[identity.Key]model_state.State{
		ledgerStateKey: model_state.NewState(ledgerStateKey, "Posted", "", ""),
	})
	ledgerClass.SetEvents(map[identity.Key]model_state.Event{})
	ledgerClass.SetGuards(map[identity.Key]model_state.Guard{})
	ledgerClass.SetActions(map[identity.Key]model_state.Action{ledgerActionKey: ledgerAction})
	ledgerClass.SetQueries(map[identity.Key]model_state.Query{})
	ledgerClass.SetTransitions(map[identity.Key]model_state.Transition{})

	model := testModel(classEntry(accountClass, accountKey), classEntry(ledgerClass, ledgerKey))
	catalog := NewClassCatalog(schema.New(model))
	PopulateDerivedAttributeCallersFromSchema(schema.New(model), catalog)

	ext := catalog.ExternalDerivedAttributes(accountKey)
	assert.Empty(t, ext, "balance referenced by simulatable ledger class should be internal")
}

func TestExternalDerivedAttributes_IncludesUncalledDerivedAttribute(t *testing.T) {
	accountKey := mustKey("domain/finance/subdomain/wallet/class/account")
	balanceAttrKey := helper.Must(identity.NewAttributeKey(accountKey, "balance"))

	balanceDeriv := model_logic.NewLogic(
		mustKey("invariant/10"),
		model_logic.LogicTypeValue,
		"Constant zero.",
		"",
		logicSpecWithExpression(&me.IntLiteral{Value: big.NewInt(0)}),
		nil,
	)
	balanceAttr := helper.Must(model_class.NewAttribute(
		balanceAttrKey,
		model_class.AttributeDetails{Name: "balance", Details: ""},
		"",
		&balanceDeriv,
		false,
		model_class.AttributeAnnotations{},
	))

	stateKey := helper.Must(identity.NewStateKey(accountKey, "open"))
	createEventKey := helper.Must(identity.NewEventKey(accountKey, "create"))
	transKey := helper.Must(identity.NewTransitionKey(accountKey, "", "create", "", "", "open"))
	accountClass := model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"})
	accountClass.SetAttributes([]model_class.Attribute{balanceAttr})
	accountClass.SetStates(map[identity.Key]model_state.State{
		stateKey: model_state.NewState(stateKey, "Open", "", ""),
	})
	accountClass.SetEvents(map[identity.Key]model_state.Event{
		createEventKey: model_state.NewEvent(createEventKey, "create", "", nil),
	})
	accountClass.SetGuards(map[identity.Key]model_state.Guard{})
	accountClass.SetActions(map[identity.Key]model_state.Action{})
	accountClass.SetQueries(map[identity.Key]model_state.Query{})
	accountClass.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: model_state.NewTransition(
			transKey,
			createEventKey,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateKey},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil},
			"",
		),
	})

	model := testModel(classEntry(accountClass, accountKey))
	catalog := NewClassCatalog(schema.New(model))
	PopulateDerivedAttributeCallersFromSchema(schema.New(model), catalog)

	ext := catalog.ExternalDerivedAttributes(accountKey)
	require.Len(t, ext, 1)
	assert.Equal(t, "balance", ext[0].Name)
}

func logicSpecWithExpression(expr me.Expression) logic_spec.ExpressionSpec {
	return logic_spec.ExpressionSpec{
		Notation:      model_logic.NotationTLAPlus,
		Specification: "stub",
		Expression:    expr,
	}
}
